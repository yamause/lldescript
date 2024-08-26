package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/openconfig/gnmic/pkg/api"
	"github.com/openconfig/gnmic/pkg/api/target"
	"google.golang.org/protobuf/encoding/prototext"
)

type Neighbor struct {
	ChassisID             string `json:"openconfig-lldp:chassis-id"`
	ChassisIDType         string `json:"openconfig-lldp:chassis-id-type"`
	ID                    string `json:"openconfig-lldp:id"`
	LastUpdateTime        string `json:"arista-lldp-augments:last-update-time"`
	ManagementAddress     string `json:"openconfig-lldp:management-address"`
	ManagementAddressType string `json:"openconfig-lldp:management-address-type"`
	PortDescription       string `json:"openconfig-lldp:port-description,omitempty"`
	PortID                string `json:"openconfig-lldp:port-id"`
	PortIDType            string `json:"openconfig-lldp:port-id-type"`
	RegistrationTime      string `json:"arista-lldp-augments:registration-time"`
	SystemDescription     string `json:"openconfig-lldp:system-description"`
	SystemName            string `json:"openconfig-lldp:system-name"`
}

type LLDPDate struct {
	Neighbor Neighbor
	LocalIf  string
}


func get(tg target.Target, ctx context.Context) ([]LLDPDate, error) {
	getReq, err := api.NewGetRequest(
		api.Path("/lldp/interfaces/interface/neighbors/neighbor/state"),
		api.Encoding("json_ietf"))
	if err != nil {
		log.Fatal(err)
	}

	// send the created gNMI GetRequest to the created target
	getResp, err := tg.Get(ctx, getReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get response from target: %w", err)
	}

	var lldpDatas []LLDPDate
	for _, notifi := range getResp.GetNotification() {
		for _, update := range notifi.Update {
			var neighbor Neighbor
			localif := update.Path.GetElem()[2].GetKey()["name"]
			err = json.Unmarshal(update.Val.GetJsonIetfVal(), &neighbor)
			if err != nil {
				return nil, fmt.Errorf("railed to unmarshal json: %v, error: %v", update.Val.GetJsonIetfVal(), err)
			}
			lldpDatas = append(lldpDatas, LLDPDate{neighbor, localif})
		}
	}
	return lldpDatas, nil
}

func set(tg target.Target, ctx context.Context, update []api.GNMIOption) (string, error) {

	// create a gNMI SetRequest
	setReq, err := api.NewSetRequest(update...)
	if err != nil {
		return "", fmt.Errorf("failed to set response from target: %w", err)
	}

	setResp, err := tg.Set(ctx, setReq)
	if err != nil {
		log.Fatal(err)
	}
	return prototext.Format(setResp), nil
}

func main() {

	// Todo: Implement argument options later
	addr := os.Args[1]
	user := os.Args[2]
	pass := os.Args[3]

	tg, err := api.NewTarget(
		// api.Name(""),
		api.Address(addr),
		api.Username(user),
		api.Password(pass),
		api.Insecure(true),
		api.SkipVerify(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = tg.CreateGNMIClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer tg.Close()

	gr, _ := get(*tg, ctx)

	counts := make(map[string]int)

	var update []api.GNMIOption

	for _, lldp := range gr {
		// 	Determine if there are multiple connections.
		path := fmt.Sprintf("/interfaces/interface[name=%s]/config/description", lldp.LocalIf)
		val := fmt.Sprintf("to:%s %s",lldp.Neighbor.SystemName, lldp.LocalIf)
		counts[lldp.LocalIf]++

		if counts[lldp.LocalIf] > 1 {
			val = "to:multiple connections"
		}
		update = append(update, api.Update(
			api.Path(path),
			api.Value(val, "json_ietf")))
	}

	r, err := set(*tg, ctx, update)

	fmt.Println(r)

	if err != nil {
		log.Fatal(err)
	}
}
