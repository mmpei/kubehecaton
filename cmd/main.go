package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog"
	"os/signal"
	"syscall"
	"github.com/mmpei/kubehecaton/version"
	"k8s.io/client-go/informers"
)

var (
	kubeconfig      string
	namespace       string
	healthCheckAddr string
	metricsAddr     string
	printVersion    bool
	resyncDuration    time.Duration

	leaderElectionFile string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&healthCheckAddr, "health-check-addr", "0.0.0.0:8080", "The address on which the health check HTTP server will listen to")
	flag.StringVar(&metricsAddr, "metrics-addr", "0.0.0.0:8086", "The address on which the prometheus metrics HTTP server will listen to")
	flag.StringVar(&namespace, "namespace", "skiff-cicd", "pod env")
	flag.BoolVar(&printVersion, "version", false, "Show version and quit")
	flag.StringVar(&leaderElectionFile, "leader-election-file", "skiff-deploy-operator", "Lock file name for leader election")
	flag.DurationVar(&resyncDuration, "resync-duration", 0, "Resync time of informer")

	flag.Parse()
}

func main() {
	if printVersion {
		fmt.Println("kubeHecaton Version:", version.Version)
		fmt.Println("Git SHA:", version.GitSHA)
		fmt.Println("Go Version:", runtime.Version())
		fmt.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	klog.Infof("kubeHecaton Version: %v", version.Version)
	klog.Infof("Git SHA: %s", version.GitSHA)
	klog.Infof("Go Version: %s", runtime.Version())
	klog.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)

	id, err := os.Hostname()
	if err != nil {
		klog.Fatalf("failed to get hostname: %v", err)
	}

	var cfg *rest.Config
	if len(kubeconfig) > 0 {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			klog.Fatalf("Error building kubeconfig: %s", err.Error())
		}
	} else {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("failed to get config: %v", err)
		}
	}

	kubeCli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("failed to get kubernetes Clientset: %v", err)
	}

	rl, err := resourcelock.New(resourcelock.EndpointsResourceLock,
		namespace,
		leaderElectionFile,
		kubeCli.CoreV1(),
		nil,
		resourcelock.ResourceLockConfig{
			Identity: id,
		})
	if err != nil {
		klog.Fatalf("error creating lock: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// catch the signal of TERM and INT
	setUpSignal(cancel)

	onStarted := func(ctx context.Context) {
		kubeInformerFactory := informers.NewSharedInformerFactory(kubeCli, resyncDuration)
		//// init Prometheus metrics provider
		//metricsProvider := utils.NewPrometheusMetricsProvider()
		//workqueue.SetProvider(metricsProvider)

		kubeInformerFactory.Start(ctx.Done())
		go func() {
			mux := http.NewServeMux()
			mux.Handle("/metrics", promhttp.Handler())
			err := http.ListenAndServe(metricsAddr, mux)
			if err != nil {
				klog.Fatal("failed to start metrics server: ", err)
			} else {
				klog.Infof("metrics server Started")
			}
		}()

		klog.Infof("controller Started")
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: onStarted,
			OnStoppedLeading: func() {
				klog.Fatalf("leader election lost")
			},
		},
	})

	panic("unreachable")
}

func setUpSignal(cancel context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		klog.Infof("received int or term, signaling shutdown")
		cancel()
	}()
}
