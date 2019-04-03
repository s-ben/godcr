package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/raedahgroup/godcr/app/sync"
	"github.com/raedahgroup/godcr/web/weblog"
)

func (routes *Routes) renderPage(tplName string, data map[string]interface{}, res http.ResponseWriter) {
	connectionInfo, err := routes.walletMiddleware.WalletConnectionInfo()
	if err != nil {
		weblog.LogError(err)
	}
	data["connectionInfo"] = connectionInfo
	routes.render(tplName, data, res)
}

func (routes *Routes) render(tplName string, data interface{}, res http.ResponseWriter) {
	if tpl, ok := routes.templates[tplName]; ok {
		err := tpl.Execute(res, data)
		if err != nil {
			log.Fatalf("error executing template: %s", err.Error())
		}
		return
	}

	log.Fatalf("template %s is not registered", tplName)
}

func (routes *Routes) renderSyncPage(syncInfo map[string]interface{}, res http.ResponseWriter) {
	syncInfo["networkType"] = routes.walletMiddleware.NetType()

	connectedPeers := syncInfo["ConnectedPeers"].(json.Number)
	if connectedPeers == "1" {
		syncInfo["ConnectedPeers"] = fmt.Sprintf("%s peer", connectedPeers)
	} else {
		syncInfo["ConnectedPeers"] = fmt.Sprintf("%s peers", connectedPeers)
	}

	if syncInfo["CurrentStep"].(json.Number) == "2" {
		// check account discovery progress percentage
		addressDiscoveryProgressString := syncInfo["AddressDiscoveryProgress"].(json.Number)
		addressDiscoveryProgress, _ := strconv.ParseInt(string(addressDiscoveryProgressString), 10, 32)
		if addressDiscoveryProgress > 100 {
			syncInfo["AddressDiscoveryProgress"] = fmt.Sprintf("%d%% (over)", addressDiscoveryProgress)
		} else {
			syncInfo["AddressDiscoveryProgress"] = fmt.Sprintf("%d%%", addressDiscoveryProgress)
		}
	}

	totalDiscoveryTimeString := syncInfo["TotalDiscoveryTime"].(json.Number)
	if totalDiscoveryTimeString != "-1" {
		totalDiscoveryTime, _ := strconv.ParseFloat(string(totalDiscoveryTimeString), 64)
		syncInfo["TotalDiscoveryTime"] = sync.CalculateTotalTimeRemaining(totalDiscoveryTime)
	}

	routes.render("sync.html", syncInfo, res)
}

func (routes *Routes) renderError(errorMessage string, res http.ResponseWriter) {
	data := map[string]interface{}{
		"error": errorMessage,
	}
	routes.renderPage("error.html", data, res)
}

func (routes *Routes) renderNoWalletError(res http.ResponseWriter) {
	data := map[string]interface{}{
		"noWallet": true,
	}
	routes.render("error.html", data, res)
}

func renderJSON(data interface{}, res http.ResponseWriter) {
	d, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("error marshalling data: %s", err.Error())
	}

	res.Header().Set("Content-Type", "application/json")
	res.Write(d)
}
