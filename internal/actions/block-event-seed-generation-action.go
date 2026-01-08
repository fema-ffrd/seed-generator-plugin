package actions

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/fema-ffrd/cc-go-sdk"
	. "github.com/fema-ffrd/seed-generator-plugin/internal/compute"
)

const (
	blockEventSeedGenerationActionName string = "block-event-seed-generation"
)

func init() {
	cc.ActionRegistry.RegisterAction(blockEventSeedGenerationActionName, &BlockEventSeedGenerationAction{})
}

type BlockEventSeedGenerationAction struct {
	cc.ActionRunnerBase
}

func (bsg *BlockEventSeedGenerationAction) Run() error {
	action := bsg.Action
	pm := bsg.PluginManager
	var eventGeneratorModel BlockModel
	eventGeneratorModel.InitialEventSeed = action.Attributes.GetInt64OrFail("initial_event_seed")
	eventGeneratorModel.InitialRealizationSeed = action.Attributes.GetInt64OrFail("initial_realization_seed")
	plugins, err := action.Attributes.GetStringSlice("plugins")
	if err != nil {
		return err
	}
	eventGeneratorModel.Plugins = plugins
	blockreader, err := action.IOManager.GetReader(cc.DataSourceOpInput{
		DataSourceName: "blockfile",
		PathKey:        "default",
	})
	if err != nil {
		return err
	}
	defer blockreader.Close()
	var blocks []Block
	err = json.NewDecoder(blockreader).Decode(&blocks)
	if err != nil {
		return err
	}

	eventNumber, err := strconv.Atoi(pm.EventIdentifier)
	if err != nil {
		return err
	}
	modelResult, err := eventGeneratorModel.Compute(eventNumber, blocks)
	if err != nil {
		return err
	}

	bytedata, err := json.Marshal(modelResult)
	if err != nil {
		return err
	}
	breader := bytes.NewReader(bytedata)
	_, err = action.Put(cc.PutOpInput{
		SrcReader: breader,
		DataSourceOpInput: cc.DataSourceOpInput{
			DataSourceName: "seeds",
			PathKey:        "default",
		},
	})
	return err
}
