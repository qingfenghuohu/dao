package dao

type ModelData struct {
	Model     ModelInfo
	Operation string
	Data      []RecordData
}

type RecordData struct {
	BeData    map[string]interface{}
	AfterData map[string]interface{}
}

func (md ModelData) GetOperation() string {
	return md.Operation
}

func (md ModelData) GetModel() ModelInfo {
	return md.Model
}

func (md *ModelData) SetData(BeData map[string]interface{}, AfterData map[string]interface{}) {
	md.Data = append(md.Data, RecordData{BeData, AfterData})
}

func NewModelData(Model ModelInfo, Operation string) ModelData {
	return ModelData{Model: Model, Operation: Operation, Data: []RecordData{}}
}
