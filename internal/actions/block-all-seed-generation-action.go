package actions

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"

	"github.com/fema-ffrd/cc-go-sdk"
	. "github.com/fema-ffrd/seed-generator-plugin/internal/compute"
)

const (
	blockAllSeedGenerationActionName string = "block-all-seed-generation"
)

func init() {
	cc.ActionRegistry.RegisterAction(blockAllSeedGenerationActionName, &BlockAllSeedGenerationAction{})
}

type BlockAllSeedGenerationAction struct {
	cc.ActionRunnerBase
}

func (bas *BlockAllSeedGenerationAction) Run() error {
	action := bas.Action
	pm := bas.PluginManager

	var eventGeneratorModel BlockModel
	eventGeneratorModel.InitialEventSeed = action.Attributes.GetInt64OrFail("initial_event_seed")
	eventGeneratorModel.InitialRealizationSeed = action.Attributes.GetInt64OrFail("initial_realization_seed")
	plugins, err := action.Attributes.GetStringSlice("plugins")

	if err != nil {
		return err
	}
	eventGeneratorModel.Plugins = plugins
	var blocks []Block
	block_name := action.Attributes.GetStringOrFail("block_dataset_name")
	blockStoreType := action.Attributes.GetStringOrFail("block_store_type")
	seed_name := action.Attributes.GetStringOrFail("seed_dataset_name")
	seedStoreType := action.Attributes.GetStringOrFail("seed_store_type")
	if blockStoreType == "eventstore" {
		recordset, err := cc.NewEventStoreRecordset(pm, &blocks, "eventstore", block_name)
		if err != nil {
			log.Fatal(err)
		}
		//err = recordset.Create()
		if err != nil {
			log.Fatal(err)
		}
		arrayResult, err := recordset.Read()
		if err != nil {
			log.Fatal(err)
		}

		for i := 0; i < arrayResult.Size(); i++ {
			var block Block
			arrayResult.Scan(&block)
			blocks = append(blocks, block)
		}
	} else {

		blockreader, err := action.IOManager.GetReader(cc.DataSourceOpInput{
			DataSourceName: block_name,
			PathKey:        "default",
		})
		if err != nil {
			return err
		}
		defer blockreader.Close()

		err = json.NewDecoder(blockreader).Decode(&blocks)
		if err != nil {
			return err
		}
	}

	modelResult, err := eventGeneratorModel.ComputeAll(blocks)
	if err != nil {
		return err
	}

	if seedStoreType == "eventstore" {
		//get the store
		ds, err := pm.GetStore(seedStoreType)
		if err != nil {
			return err
		}
		//check if it is the right store type
		tdbds, ok := ds.Session.(cc.MultiDimensionalArrayStore)
		if ok {
			arrayInput := cc.CreateArrayInput{
				ArrayPath: seed_name,
				Dimensions: []cc.ArrayDimension{
					{
						Name:          "Events", //row
						DimensionType: cc.DIMENSION_INT,
						Domain:        []int64{1, int64(len(modelResult))},
						TileExtent:    1,
					}, {
						Name:          "Plugins", //column
						DimensionType: cc.DIMENSION_INT,
						Domain:        []int64{1, int64(len(plugins))},
						TileExtent:    int64(len(plugins)),
					}, /*{
						Name: "Seeds",
						DimensionType: cc.DIMENSION_INT,
						Domain: []int64{1,2},
						TileExtent: 1,//what does this mean really.
					}*/
				},
				Attributes: []cc.ArrayAttribute{
					{Name: "realization_seed", DataType: cc.ATTR_INT64},
					{Name: "block_seed", DataType: cc.ATTR_INT64},
					{Name: "event_seed", DataType: cc.ATTR_INT64},
				},
				ArrayType:  cc.ARRAY_DENSE,
				CellLayout: cc.ROWMAJOR,
				TileLayout: cc.COLMAJOR,
			}
			err = tdbds.CreateArray(arrayInput)
			if err != nil {
				return err
			}

			//now make a put array input and put the data properly arranged.
			eventseeddata := make([]int64, (len(plugins))*(len(modelResult)))
			blockseeddata := make([]int64, (len(plugins))*(len(modelResult)))
			realizationseeddata := make([]int64, len(plugins)*len(modelResult))
			for i, ec := range modelResult {
				for j, plugin := range plugins {
					ss := ec.Seeds[plugin]
					eventseeddata[(i*len(plugins))+j] = ss.EventSeed
					blockseeddata[(i*len(plugins))+j] = ss.BlockSeed
					realizationseeddata[(i*len(plugins))+j] = ss.RealizationSeed
				}
			}
			//create a buffer
			buffer := []cc.PutArrayBuffer{
				{
					AttrName: "event_seed",
					Buffer:   eventseeddata,
				}, {
					AttrName: "block_seed",
					Buffer:   blockseeddata,
				},
				{
					AttrName: "realization_seed",
					Buffer:   realizationseeddata,
				},
			}
			//create an input
			input := cc.PutArrayInput{
				Buffers:   buffer,
				DataPath:  seed_name,
				ArrayType: cc.ARRAY_DENSE,
			}
			err = tdbds.PutArray(input)
			if err != nil {
				return err
			}
			mds, ok := ds.Session.(cc.MetadataStore)
			if ok {
				err = mds.PutMetadata("seed_columns", plugins)
				if err != nil {
					return err
				}

			} else {
				return errors.New("could not store metadata on which plugins recieve seeds")
			}
		} else {
			//store does not support this data.
			return errors.New("store session is not a multidimensional array store")
		}
	} else {
		bytedata, err := json.Marshal(modelResult)
		if err != nil {
			return err
		}
		breader := bytes.NewReader(bytedata)
		_, err = action.Put(cc.PutOpInput{
			SrcReader: breader,
			DataSourceOpInput: cc.DataSourceOpInput{
				DataSourceName: seed_name,
				PathKey:        "default",
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
