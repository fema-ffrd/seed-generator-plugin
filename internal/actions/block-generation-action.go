package actions

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/fema-ffrd/cc-go-sdk"
	. "github.com/fema-ffrd/seed-generator-plugin/internal/compute"
)

const (
	blockGenFixedLengthName string = "block-generation-fixed-length"
	blockGenerationName     string = "block-generation"
	defaultSeed             int64  = 1234
)

func init() {
	cc.ActionRegistry.RegisterAction(blockGenerationName, &BlockGenerationAction{})
	cc.ActionRegistry.RegisterAction(blockGenFixedLengthName, &BlockGenerationAction{})
}

type BlockGenerationAction struct {
	cc.ActionRunnerBase
}

func (bga *BlockGenerationAction) Run() error {
	action := bga.Action
	pm := bga.PluginManager
	fixedLength := false
	if action.Name == blockGenFixedLengthName {
		fixedLength = true
	}

	blockGenerator := BlockGenerator{
		TargetTotalEvents:    action.Attributes.GetInt64OrFail("target_total_events"),
		BlocksPerRealization: action.Attributes.GetIntOrFail("blocks_per_realization"),
		TargetEventsPerBlock: action.Attributes.GetIntOrFail("target_events_per_block"),
		Seed:                 action.Attributes.GetInt64OrDefault("seed", defaultSeed),
	}

	var blocks []Block
	if fixedLength {
		blocks = blockGenerator.GenerateBlocksFixedLength()
	} else {
		blocks = blockGenerator.GenerateBlocks()
	}

	bytedata, err := json.Marshal(blocks)
	if err != nil {
		return fmt.Errorf("could not encode blocks, %v", err)
	}
	outputDataset_name := action.Attributes.GetStringOrFail("outputdataset_name")
	storeType := action.Attributes.GetStringOrFail("store_type")
	if storeType == "eventstore" {
		recordset, err := cc.NewEventStoreRecordset(pm, &blocks, "eventstore", outputDataset_name)
		if err != nil {
			bga.Log("failed to get a new event store recordset", "error", err)
			return err
		}
		err = recordset.Create()
		if err != nil {
			bga.Log("failed to create the new event store recordset", "error", err)
			return err
		}
		err = recordset.Write(&blocks)
		if err != nil {
			bga.Log("failed to write the block recordeset", "error", err)
			return err
		}
	} else {
		breader := bytes.NewReader(bytedata)
		_, err = pm.Put(cc.PutOpInput{
			SrcReader: breader,
			DataSourceOpInput: cc.DataSourceOpInput{
				DataSourceName: outputDataset_name,
				PathKey:        "default",
			},
		})
		if err != nil {
			bga.Log("failed to copy the block file to the remote store", "error", err)
			return err
		}
	}
	return nil
}
