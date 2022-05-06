package views

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/gravitl/netmaker/netclient/config"
	"github.com/gravitl/netmaker/netclient/functions"
	"github.com/gravitl/netmaker/netclient/gui/components"
	"github.com/gravitl/netmaker/netclient/ncutils"
)

var currentNetwork *string

// GetNetworksView - displays the view of all networks
func GetNetworksView(networks []string) fyne.CanvasObject {
	// renders := []fyne.CanvasObject{}
	if networks == nil || len(networks) == 0 {
		return container.NewCenter(widget.NewLabel("No networks present"))
	}
	grid := container.New(layout.NewGridLayout(4),
		container.NewCenter(widget.NewLabel("Network Name")),
		container.NewCenter(widget.NewLabel("Node Info")),
		container.NewCenter(widget.NewLabel("Pull Latest")),
		container.NewCenter(widget.NewLabel("Leave network")),
	)
	for i := range networks {
		network := &networks[i]
		grid.AddObject(
			container.NewCenter(widget.NewLabel(*network)),
		)
		grid.AddObject(
			components.ColoredIconButton("info", theme.InfoIcon(), func() {
				RefreshComponent(NetDetails, GetSingleNetworkView(*network))
				ShowView(NetDetails)
			}, components.Gold_color),
		)
		grid.AddObject(
			components.ColoredIconButton("pull", theme.DownloadIcon(), func() {
				// TODO call pull with network name
				pull(*network)
			}, components.Blue_color),
		)
		grid.AddObject(
			components.ColoredIconButton("leave", theme.DeleteIcon(), func() {
				leave(*network)
			}, components.Danger_color),
		)
		// renders = append(renders, container.NewCenter(netToolbar))
	}

	return container.NewCenter(grid)
}

const fakeData = `{"networks":[{"name":"devops","node_id":"5aeb3e18-1236-46d8-8415-8699bfe5d44e","current_node":{"name":"ingress","interface":"nm-devops","private_ipv4":"10.10.10.1","public_endpoint":"167.71.106.39"},"peers":[{"public_key":"QlLJlQKy6C7XirHdnkXiMcCSCed2ieDt6qL3DSzjSxo=","public_endpoint":"167.71.100.69:51821","addresses":[{"cidr":"10.10.10.3/32","ip":"10.10.10.3"},{"cidr":"10.10.10.0/24","ip":"10.10.10.0"}]},{"public_key":"WnU5t2Rl9kD7lzASe8nH7VyS+jhTLUCigMJKKt+UrnU=","public_endpoint":"167.71.98.164:51821","addresses":[{"cidr":"10.10.10.2/32","ip":"10.10.10.2"},{"cidr":"165.227.116.94/32","ip":"165.227.116.94"}]},{"public_key":"rRI9qNHIiSQsIyZgGBvyZML98bZ6z8iZYfZLWPSZJ1k=","public_endpoint":"167.71.100.25:51821","addresses":[{"cidr":"10.10.10.5/32","ip":"10.10.10.5"}]},{"public_key":"R7JoXHCj9q/yXizr9q7p3xW5dxAX+l6Hg17k/98T0GI=","public_endpoint":"167.71.164.7:51821","addresses":[{"cidr":"10.10.10.254/32","ip":"10.10.10.254"}]},{"public_key":"M5gwhvr1Qrg55gGrPrkd3NbLJoDqTsjiEPvvf1yyaiQ=","public_endpoint":"\u003cnil\u003e","addresses":[{"cidr":"10.10.10.6/32","ip":"10.10.10.6"}]}]}]}`

// GetSingleNetworkView - returns details and option to pull a network
func GetSingleNetworkView(network string) fyne.CanvasObject {
	if network == "" || len(network) == 0 {
		return container.NewCenter(widget.NewLabel("No valid network selected"))
	}

	// == read node values ==
	LoadingNotify()
	nets, err := functions.List(network)
	if err != nil || len(nets) < 1 {
		return container.NewCenter(widget.NewLabel("No data retrieved."))
	}
	var nodecfg config.ClientConfig
	nodecfg.Network = network
	nodecfg.ReadConfig()
	nodeID := nodecfg.Node.ID
	lastCheckInTime := time.Unix(nodecfg.Node.LastCheckIn, 0)
	lastCheckIn := lastCheckInTime.Format("2006-01-02 15:04:05")
	privateAddr := nodecfg.Node.Address
	privateAddr6 := nodecfg.Node.Address6
	endpoint := nodecfg.Node.Endpoint
	health := " (HEALTHY)"
	if time.Now().After(lastCheckInTime.Add(time.Minute * 5)) {
		health = " (WARNING)"
	} else if time.Now().After(lastCheckInTime.Add(time.Minute * 30)) {
		health = " (ERROR)"
	}
	lastCheckIn += health
	version := nodecfg.Node.Version
	peerString := ""
	for _, peer := range nets[0].Peers {
		peerString += fmt.Sprintf("- Endpoint: %s, Addresses:", peer.PublicEndpoint)
		for _, addr := range peer.Addresses {
			peerString += fmt.Sprintf(", %s", addr.IP)
		}
		peerString += "\n"
	}
	pullBtn := components.ColoredButton("pull "+network, func() { pull(network) }, components.Blue_color)
	pullBtn.Resize(fyne.NewSize(pullBtn.Size().Width, 50))
	netDetailsView := container.NewCenter(
		// components.ColoredText("Selected "+network, components.Orange_color),
		container.NewGridWithColumns(1, widget.NewRichTextFromMarkdown(fmt.Sprintf(`### %s
- ID: %s
- Last Check In: %s
- Endpoint: %s
- Address (IPv4): %s
- Address6 (IPv6): %s
- Version: %s
### Peers
%s
`, network, nodeID, lastCheckIn, endpoint, privateAddr, privateAddr6, version, peerString)),
			container.NewCenter(pullBtn),
		))
	ClearNotification()
	return netDetailsView
}

// == private ==
func pull(network string) {
	LoadingNotify()
	_, err := functions.Pull(network, true)
	if err != nil {
		ErrorNotify("Failed to pull " + network + " : " + err.Error())
	} else {
		SuccessNotify("Pulled " + network + "!")
	}
}

func leave(network string) {

	confirmView := GetConfirmation("Confirm leaving "+network+"?", func() {
		ShowView(Networks)
	}, func() {
		LoadingNotify()
		err := functions.LeaveNetwork(network, true)
		if err != nil {
			ErrorNotify("Failed to leave " + network + " : " + err.Error())
		} else {
			SuccessNotify("Left " + network)
		}
		networks, err := ncutils.GetSystemNetworks()
		if err != nil {
			networks = []string{}
			ErrorNotify("Failed to read local networks!")
		}
		RefreshComponent(Networks, GetNetworksView(networks))
		ShowView(Networks)
	})
	RefreshComponent(Confirm, confirmView)
	ShowView(Confirm)
}