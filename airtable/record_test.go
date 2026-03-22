package airtable


import (
	"testing"

	. "github.com/Antfood/airgo/testutils/testutils"
)


type testRecSchema struct {
   Id string `json:"id"`
   Company string `json:"company"`
   Address string `json:"address"`
}

func TestRecords(t *testing.T){
   t.Run("NewRecord", testNewRecord)
}


func testNewRecord(t *testing.T){

   record := NewRecord[testRecSchema]("base_id","table_id")

   Assert(t, record.TableId == "table_id", "Expected 'table_id', got '%s'", record.TableId)
   Assert(t, record.BaseId == "base_id", "Expected 'base_id', got '%s'", record.BaseId)
}
