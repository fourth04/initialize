package sysinfo

import "testing"

func TestSaveConfig(t *testing.T) {
	ifcfg := newDefaultIfCfg()
	err := ifcfg.SaveConfigFile("./test.cfg")
	if err != nil {
		t.Error("保存失败！")
	}
	t.Log("保存成功！")
}
