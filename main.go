package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog"
)

const (
	Once   = "once"
	Always = "always"
)

type config struct {
	directory        string
	mode             string // once | always
	nodeName         string
	kubeconfig       string
	deleteStaleFiles bool
}

type nodeLabelsToFiles struct {
	config    *config
	clientset *kubernetes.Clientset
}

func (c *config) Validate() error {
	if c.directory == "" {
		return errors.New("directory is not configured")
	}
	if c.directory == "" {
		return errors.New("directory is not configured")
	}
	if c.nodeName == "" {
		return errors.New("nodename is not configured")
	}
	if c.mode == "" {
		return errors.New("mode is not configured")
	}
	if (c.mode != Once) && (c.mode != Always) {
		return fmt.Errorf("mode should be one of %s or %s", Once, Always)
	}
	return nil
}

func (n *nodeLabelsToFiles) parseFlags() error {
	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(
			&(n.config.kubeconfig), "kubeconfig",
			filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file. Can be overridden via "+
				"the environment variable KUBECONFIG")
	} else {
		flag.StringVar(&(n.config.kubeconfig), "kubeconfig", "",
			"absolute path to the kubeconfig file. Can be overridden via the "+
				"environment variable KUBECONFIG")
	}
	flag.BoolVar(&(n.config.deleteStaleFiles), "delete-stale-files", true,
		"This determines if node-labels-to-path will delete stale files or "+
			"files it is not aware of or keep them, by default it will delete "+
			"them. Can be overriden via the environment variable DELETE_STALE_FILES")
	flag.StringVar(&(n.config.directory), "directory", "", "Directory to "+
		"write the node labels in, if the directory does not exist "+
		"node-labels-to-files will create it. Can be overridden via the "+
		"environment variable DIRECTORY")
	flag.StringVar(&(n.config.mode), "mode", Always, "This determines "+
		"the mode n works in, when it is set to once it retrieves the node "+
		"labels and exits, if set to always it creates a watch on the node and "+
		"will detect and update the directory to reflect the labels when they "+
		"change on the node. Acceptable options is either of "+Always+"|"+Once+
		"Can be overriden via the environment variable MODE")
	flag.StringVar(&(n.config.nodeName), "nodename", "", "Name of node "+
		"whose label n should retrieve. Can be overridden via the environment"+
		"variable NODENAME")
	klog.InitFlags(nil)
	flag.Parse()
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		n.config.kubeconfig = kubeconfig
	}
	if nodeName := os.Getenv("NODENAME"); nodeName != "" {
		n.config.nodeName = nodeName
	}
	if directory := os.Getenv("DIRECTORY"); directory != "" {
		n.config.directory = directory
	}
	if mode := os.Getenv("MODE"); mode != "" {
		n.config.mode = mode
	}
	if deleteStaleFiles := os.Getenv("DELETE_STALE_FILES"); deleteStaleFiles != "" {
		if deleteStaleFiles == "false" {
			n.config.deleteStaleFiles = false
		}
	}
	return n.config.Validate()
}

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (n *nodeLabelsToFiles) getConfig() (*rest.Config, error) {
	if n.config.kubeconfig == "" || !Exists(n.config.kubeconfig) {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", n.config.kubeconfig)
}

func (n *nodeLabelsToFiles) buildClient() {
	config, err := n.getConfig()
	if err != nil {
		klog.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}
	n.clientset = clientset
}

func (n *nodeLabelsToFiles) filesToDelete(labels map[string]string) ([]string,
	error) {
	var files []string
	labelsToPaths := make(map[string]string)

	for label, value := range labels {
		labelWithPath := filepath.Join(n.config.directory, label)
		labelsToPaths[labelWithPath] = value
		dir := path.Dir(labelWithPath)
		for dir != strings.TrimSuffix(n.config.directory, "/") {
			labelsToPaths[dir] = ""
			dir = path.Dir(dir)
		}
	}

	err := filepath.Walk(
		n.config.directory,
		func(path string, info os.FileInfo, err error) error {
			if _, ok := labelsToPaths[path]; path != n.config.directory && !ok {
				files = append(files, path)
			}
			return nil
		})

	return files, err
}

func (n *nodeLabelsToFiles) deleteStaleFiles(labels map[string]string) error {
	files, err := n.filesToDelete(labels)
	if err != nil {
		return err
	}

	for _, file := range files {
		klog.V(2).Info("Removing stale file: ", file)
		os.RemoveAll(file)
	}

	return nil
}

func (n *nodeLabelsToFiles) createFileFromLabels(labels map[string]string) {
	for fileName, fileContent := range labels {
		err := writeToFile(filepath.Join(n.config.directory, fileName),
			fileContent)
		klog.V(5).Infof("Creating file: %s, with content: %s", fileName, fileContent)
		if err != nil {
			klog.Errorf("Failed creating file: %s, due to: %s", fileName, err)
		}
	}
}

func writeToFile(fileName string, fileContent string) error {
	dir := filepath.Dir(fileName)
	dirInfo, err := os.Stat(dir)
	if err != nil || !dirInfo.IsDir() {
		klog.Infof("Creating sub-directory: %s", dir)
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(fileContent))
	klog.V(2).Infof("Writing: '%s' in file: %s", fileContent, fileName)
	if err != nil {
		return err
	}
	return nil
}

func (n *nodeLabelsToFiles) getNodeLabels() (map[string]string, error) {
	node, err := n.clientset.CoreV1().Nodes().Get(n.config.nodeName,
		metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return node.GetLabels(), nil
}

func (n *nodeLabelsToFiles) processOnce(labels map[string]string) {
	klog.V(2).Info("Refreshing Labels information for: ", n.config.nodeName)
	klog.V(2).Infof("Retrieved labels: %v for node: %s", labels,
		n.config.nodeName)
	n.createFileFromLabels(labels)
	var err error
	if n.config.deleteStaleFiles {
		err = n.deleteStaleFiles(labels)
	} else {
		klog.Info("Skipping deletion of stale files")
	}
	if err != nil {
		klog.Fatalf("Encountered: %s, while trying to delete stale files with "+
			"labels: %v", err, labels)
	}
}

func (n *nodeLabelsToFiles) process(closeChan <-chan string) {
	nodeWatch, err := n.clientset.CoreV1().Nodes().Watch(
		metav1.ListOptions{
			FieldSelector: "metadata.name=" + n.config.nodeName,
		})

	if err != nil {
		klog.Fatalf("Encountered: %s, while trying to establish a watch on: %s",
			err, n.config.nodeName)
	}
	watchChan := nodeWatch.ResultChan()

	for {
		select {
		case <-closeChan:
			nodeWatch.Stop()
			return
		case ev := <-watchChan:
			labels := ev.Object.(*v1.Node).GetLabels()
			n.processOnce(labels)
		}
	}
}

func main() {
	nodeLabelsToFiles := &nodeLabelsToFiles{config: &config{}}
	if err := nodeLabelsToFiles.parseFlags(); err != nil {
		klog.Fatalf("Missing configuration: %s", err)
	}
	nodeLabelsToFiles.buildClient()
	klog.Info("Starting node-labels-to-files")
	labels, err := nodeLabelsToFiles.getNodeLabels()
	if err != nil {
		klog.Fatalf("Error retrieving labels: %s", err)
	}
	nodeLabelsToFiles.processOnce(labels)
	if nodeLabelsToFiles.config.mode == Always {
		closeChan := make(chan string)
		defer close(closeChan)
		nodeLabelsToFiles.process(closeChan)
	}
}
