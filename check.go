package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// WeChat notification function
func sendWechatNotification(content string) {
	url := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=0221b04a-c5d6-46f7-ae67-29eb8fd11620"
	jsonData := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}
	jsonValue, _ := json.Marshal(jsonData)
	_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Println("Error sending WeChat notification:", err)
	}
}

// Get node region and cluster info based on IP address
func getRegion(ip string) (string, string, string, string) {
	formattedIP := strings.ReplaceAll(ip, ".", "-")
	nodename := fmt.Sprintf("host-%s-%s", formattedIP, ip)
	info, err := exec.Command("grep", "-r", nodename, "/data/cluster_node_info/").Output()
	if err != nil {
		fmt.Println("Error: Node not found in any K8s cluster")
		return "", "", "", ""
	}
	infoParts := strings.Fields(string(info))
	if len(infoParts) < 5 {
		return "", "", "", ""
	}
	return infoParts[1], infoParts[2], infoParts[3], infoParts[4]
}

// Cordon node and set maintenance label
func cordonNode(ip, az, cluster string) {
	formattedIP := strings.ReplaceAll(ip, ".", "-")
	nodename := fmt.Sprintf("host-%s-%s", formattedIP, ip)
	kubeconfig := fmt.Sprintf("/home/ubuntu/.kube/%s-%s.kubeconfig", az, cluster)
	cmd := exec.Command("kubectl", "--kubeconfig="+kubeconfig, "cordon", nodename)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error cordoning node:", err)
		return
	}
	cmd = exec.Command("/home/opbin/n9e_alarm_operator.sh", "-i", ip, "-a", "block")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error blocking alarm:", err)
		return
	}
	cmd = exec.Command("kubectl", "--kubeconfig="+kubeconfig, "label", "node", nodename, "belong-usage-type=repair", "--overwrite")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error labeling node:", err)
		return
	}
}

// Uncordon node and remove maintenance label
func uncordonNode(ip, az, cluster string) {
	formattedIP := strings.ReplaceAll(ip, ".", "-")
	nodename := fmt.Sprintf("host-%s-%s", formattedIP, ip)
	kubeconfig := fmt.Sprintf("/home/ubuntu/.kube/%s-%s.kubeconfig", az, cluster)
	cmd := exec.Command("kubectl", "--kubeconfig="+kubeconfig, "uncordon", nodename)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error uncordoning node:", err)
		return
	}
	cmd = exec.Command("/home/opbin/n9e_alarm_operator.sh", "-i", ip, "-a", "clear")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error clearing alarm:", err)
		return
	}
	cmd = exec.Command("kubectl", "--kubeconfig="+kubeconfig, "label", "node", nodename, "belong-usage-type-")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error removing label:", err)
		return
	}
}

// Process IP or IP list file for maintenance
func processMaintenance(param, message string) {
	var ips []string
	if strings.Contains(param, ".") {
		ips = append(ips, param)
	} else {
		file, err := os.Open(param)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			ips = append(ips, scanner.Text())
		}
	}

	var wg sync.WaitGroup
	for _, ip := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			region, az, cluster, _ := getRegion(ip)
			if region == "" {
				return
			}
			cordonNode(ip, az, cluster)
		}(ip)
	}
	wg.Wait()

	user, err := exec.Command("who", "-m").Output()
	if err != nil {
		fmt.Println("Error getting user:", err)
		return
	}
	content := fmt.Sprintf("Maintenance User: %s\nAffected Nodes:\n%s\nReason: %s", strings.TrimSpace(string(user)), strings.Join(ips, "\n"), message)
	sendWechatNotification(content)
}

// List node info and check pods
func listNodeInfo(param string, checkFunc func(string, string, string)) {
	var ips []string
	if strings.Contains(param, ".") {
		ips = append(ips, param)
	} else {
		file, err := os.Open(param)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			ips = append(ips, scanner.Text())
		}
	}

	for _, ip := range ips {
		region, az, cluster, _ := getRegion(ip)
		if region == "" {
			continue
		}
		checkFunc(ip, az, cluster)
	}
}

// Check for user pods on node
func checkPods(ip, az, cluster string) {
	formattedIP := strings.ReplaceAll(ip, ".", "-")
	nodename := fmt.Sprintf("host-%s-%s", formattedIP, ip)
	kubeconfig := fmt.Sprintf("/home/ubuntu/.kube/%s-%s.kubeconfig", az, cluster)
	cmd := exec.Command("kubectl", "--kubeconfig="+kubeconfig, "get", "pods", "-A", "--field-selector", "spec.nodeName="+nodename)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting pods:", err)
		return
	}
	if strings.Contains(string(output), "ns") {
		fmt.Printf("User pods on node %s:\n%s\n", nodename, string(output))
	} else {
		fmt.Printf("No user pods found on node %s\n", nodename)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./nrepair [OPTIONS]")
		fmt.Println("Options:")
		fmt.Println("  -h            Display this help message")
		fmt.Println("  -i IP        Input single IP or file with IP list for maintenance")
		fmt.Println("  -r IP        Recover single IP or file with IP list")
		fmt.Println("  -m MSG       Maintenance message")
		fmt.Println("  -l IP        List node name for given IP")
		return
	}

	var inputParam, maintenanceMsg string
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-h":
			fmt.Println("Usage: ./nrepair [OPTIONS]")
			fmt.Println("Options:")
			fmt.Println("  -h            Display this help message")
			fmt.Println("  -i IP        Input single IP or file with IP list for maintenance")
			fmt.Println("  -r IP        Recover single IP or file with IP list")
			fmt.Println("  -m MSG       Maintenance message")
			fmt.Println("  -l IP        List node name for given IP")
			return
		case "-i":
			i++
			inputParam = os.Args[i]
		case "-r":
			i++
			listNodeInfo(os.Args[i], uncordonNode)
		case "-m":
			i++
			maintenanceMsg = os.Args[i]
		case "-l":
			i++
			listNodeInfo(os.Args[i], checkPods)
		default:
			fmt.Println("Error: Invalid option", os.Args[i])
			return
		}
	}

	if inputParam != "" {
		processMaintenance(inputParam, maintenanceMsg)
	}
}
