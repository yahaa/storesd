package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/yahaa/storesd/utils"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"sigs.k8s.io/yaml"
)

var (
	vp      *viper.Viper
	cfgPath string
)

func init() {
	flagSet := flag.CommandLine
	klog.InitFlags(flagSet)

	flagSet.StringVar(&cfgPath, "config-path", "hack/config-local.yaml", "config path")

	flagSet.Parse(os.Args[1:])
}

func initConfig() error {
	vp = viper.New()
	vp.AddConfigPath(strings.TrimSuffix(cfgPath, filepath.Base(cfgPath)))
	vp.SetConfigName(filepath.Base(cfgPath))
	vp.SetConfigType(strings.TrimPrefix(filepath.Ext(cfgPath), "."))

	if err := vp.ReadInConfig(); err != nil {
		return err
	}

	go func() {
		vp.WatchConfig()
		vp.OnConfigChange(func(event fsnotify.Event) {
			klog.Infof("config update current targets: %v", vp.Get("syncTargets"))
		})
	}()

	return nil
}

func serve(srv *http.Server) func() error {
	return func() error {
		klog.Infof("start http server listen on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}

func doSync(target SyncTarget) (targets []string, err error) {
	client, err := utils.NewClientset(target.KubeConfigPath)
	if err != nil {
		return nil, err
	}

	for _, pair := range target.Services {
		// 获取与 service 同名的 endpoints
		eps, err := client.CoreV1().Endpoints(target.Namespace).Get(pair.Service, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("get %s/%s/%s endpoint err: %v", target.Cluster, target.Namespace, pair.Service, err)
		}

		for _, subset := range eps.Subsets {
			for _, port := range subset.Ports {
				if port.Name == pair.PortName {
					for _, addr := range subset.Addresses {
						target := fmt.Sprintf("%s:%d", addr.IP, port.Port)
						targets = append(targets, target)
					}
				}
			}
		}
	}

	return targets, nil

}

func syncService(ctx context.Context) {
	var runtimeCfg Config
	if err := vp.Unmarshal(&runtimeCfg); err != nil {
		klog.Fatalf("get runtime config err: %w", err)
	}

	klog.Infof("start sync %v to %s", runtimeCfg.SyncTargets, runtimeCfg.OutputPath)

	var oldEps []ServiceEndpoints
	outputFile := path.Join(runtimeCfg.OutputPath, "stores.yaml")

	data, err := ioutil.ReadFile(outputFile)
	if err != nil {
		klog.Infof("read old data err: %v ,skip read old data", err)
	} else {
		if err := yaml.Unmarshal(data, &oldEps); err != nil {
			klog.Infof("unmarshal data err: %v ,skip read old data")
		}
	}

	t := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-t.C:
			var runtimeCfg Config

			if err := vp.Unmarshal(&runtimeCfg); err != nil {
				klog.Errorf("get runtime config err: %w", err)
				continue
			}

			var targets []string
			for _, st := range runtimeCfg.SyncTargets {
				ts, err := doSync(st)
				if err != nil {
					klog.Errorf("do sync %s/%s/%v err: %v", st.Cluster, st.Namespace, st.Services, err)
					break
				}

				targets = append(targets, ts...)
			}

			targets = utils.StringSet(targets)

			newEps := []ServiceEndpoints{{targets}}

			skip := true
			if len(oldEps) == len(newEps) {
				for i := 0; i < len(newEps); i++ {
					if !utils.StringSliceEqual(oldEps[i].Targets, newEps[i].Targets) {
						skip = false
					}
				}
			} else {
				skip = false
			}

			if skip {
				continue
			}

			data, err := yaml.Marshal(newEps)
			if err != nil {
				klog.Errorf("yaml marshal err: %v", err)
				continue
			}

			swp := fmt.Sprintf("%s.swp", outputFile)
			if err := ioutil.WriteFile(swp, data, 0751); err != nil {
				klog.Errorf("write file %s err: %w", swp, err)
				continue
			}

			if err := os.Rename(swp, outputFile); err != nil {
				klog.Errorf("rename %s to %s err: %w", swp, outputFile, err)
				continue
			}

			oldEps = newEps

			klog.Infof("write to %s success.", outputFile)
		case <-ctx.Done():
			klog.Infof("stop sync service.")
			return
		}
	}

}

func run(ctx context.Context) func() error {
	return func() error {
		go syncService(ctx)
		return nil
	}
}

// Main 主逻辑入口
func Main() int {
	var (
		route       = gin.New()
		ctx, cancel = context.WithCancel(context.Background())
		term        = make(chan os.Signal)
		srv         = &http.Server{
			Addr:    vp.GetString("srvAddr"),
			Handler: route,
		}
	)

	route.Use(gin.Recovery())
	route.GET("/ping", ok)

	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(serve(srv))

	wg.Go(run(ctx))

	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	select {
	case <-term:
		klog.Info("received SIGTERM, exiting gracefully...")
	case <-ctx.Done():
	}

	if err := srv.Shutdown(ctx); err != nil {
		klog.Error("server shutdown err", err)
		return 1
	}

	cancel()
	if err := wg.Wait(); err != nil {
		klog.Error("unhandled error received. Exiting...", err)
		return 1
	}

	return 0
}

// ok 返回一个 ok 的状态
func ok(_ *gin.Context) {
}

func main() {
	if err := initConfig(); err != nil {
		klog.Error(err)
		os.Exit(1)
	}

	os.Exit(Main())
}
