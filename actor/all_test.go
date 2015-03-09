package actor

import (
	"encoding/json"
	"github.com/nicholaskh/assert"
	"github.com/nicholaskh/golib/breaker"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestConsts(t *testing.T) {
	assert.Equal(t, int64(1), ticket_user)
	assert.Equal(t, int64(2), ticket_alliance)
	assert.Equal(t, int64(3), ticket_chat_room)
}

func TestPushUnmarshal(t *testing.T) {
	body := `r|140102165321621504,140105095412064256,140113146319867904,139420506066653184|140246270634758144|{"cmd":"march","payload":{"data":{"uid":45,"march_id":52,"city_id":45,"opp_uid":2,"world_id":1,"type":"attack","start_x":312,"start_y":94,"end_x":0,"end_y":20,"start_time":1422272310,"end_time":1422272630,"speed":1,"state":"marching"},"user_info":[{"uid":45,"alliance_id":42,"power":11425,"name":"player45"},{"uid":2,"alliance_id":0,"power":200000,"name":"UnitTestAccount2"}],"alliance":[{"alliance_id":42,"name":"zcL8e","acronym":"zzW","level":1,"symbol_code":0}]},"time":"1422272310157","data":{"set":{"march":[{"uid":45,"march_id":52,"city_id":45,"opp_uid":2,"opp_city_id":2,"rally_id":0,"world_id":1,"type":"attack","start_x":312,"start_y":94,"end_x":0,"end_y":20,"start_time":1422272310,"end_time":1422272630,"speed":1,"start_troops":{"infantry_t1":20,"cavalry_t1":100},"troops":{"infantry_t1":20,"cavalry_t1":100},"opp_start_troops":[],"opp_troops":[],"troops_sentto_hospital":[],"resources":[],"rewards":[],"ctime":1422272310,"mtime":1422272310,"state":"marching","battle_log":[]}],"user_city":[{"city_troop":{"infantry_t1":87,"cavalry_t1":0,"ranged_t1":"100","artillery_close_t1":"100","artillery_distance_t1":"100"},"uid":45,"city_id":45}]}}}`
	p := new(Push)
	p.Body = []byte(body)
	msg, fromId, toIds := p.Unmarshal()
	assert.Equal(t, int64(140246270634758144), fromId)
	assert.Equal(t, []string{
		"140102165321621504",
		"140105095412064256",
		"140113146319867904",
		"139420506066653184",
	}, toIds)
	assert.Equal(t, true, strings.HasPrefix(msg, `{"cmd":"march","payload":{"data":{`))
}

func TestPushType(t *testing.T) {
	body := `p|140102165321621504,140105095412064256,140113146319867904,139420506066653184|140246270634758144|{"cmd":"march","payload":{"data":{"uid":45,"march_id":52,"city_id":45,"opp_uid":2,"world_id":1,"type":"attack","start_x":312,"start_y":94,"end_x":0,"end_y":20,"start_time":1422272310,"end_time":1422272630,"speed":1,"state":"marching"},"user_info":[{"uid":45,"alliance_id":42,"power":11425,"name":"player45"},{"uid":2,"alliance_id":0,"power":200000,"name":"UnitTestAccount2"}],"alliance":[{"alliance_id":42,"name":"zcL8e","acronym":"zzW","level":1,"symbol_code":0}]},"time":"1422272310157","data":{"set":{"march":[{"uid":45,"march_id":52,"city_id":45,"opp_uid":2,"opp_city_id":2,"rally_id":0,"world_id":1,"type":"attack","start_x":312,"start_y":94,"end_x":0,"end_y":20,"start_time":1422272310,"end_time":1422272630,"speed":1,"start_troops":{"infantry_t1":20,"cavalry_t1":100},"troops":{"infantry_t1":20,"cavalry_t1":100},"opp_start_troops":[],"opp_troops":[],"troops_sentto_hospital":[],"resources":[],"rewards":[],"ctime":1422272310,"mtime":1422272310,"state":"marching","battle_log":[]}],"user_city":[{"city_troop":{"infantry_t1":87,"cavalry_t1":0,"ranged_t1":"100","artillery_close_t1":"100","artillery_distance_t1":"100"},"uid":45,"city_id":45}]}}}`
	p := new(Push)
	p.Body = []byte(body)
	assert.Equal(t, "p", p.Type())
}

func BenchmarkPushType(b *testing.B) {
	b.ReportAllocs()
	body := `p|140102165321621504,140105095412064256,140113146319867904,139420506066653184|140246270634758144|{"cmd":"march","payload":{"data":{"uid":45,"march_id":52,"city_id":45,"opp_uid":2,"world_id":1,"type":"attack","start_x":312,"start_y":94,"end_x":0,"end_y":20,"start_time":1422272310,"end_time":1422272630,"speed":1,"state":"marching"},"user_info":[{"uid":45,"alliance_id":42,"power":11425,"name":"player45"},{"uid":2,"alliance_id":0,"power":200000,"name":"UnitTestAccount2"}],"alliance":[{"alliance_id":42,"name":"zcL8e","acronym":"zzW","level":1,"symbol_code":0}]},"time":"1422272310157","data":{"set":{"march":[{"uid":45,"march_id":52,"city_id":45,"opp_uid":2,"opp_city_id":2,"rally_id":0,"world_id":1,"type":"attack","start_x":312,"start_y":94,"end_x":0,"end_y":20,"start_time":1422272310,"end_time":1422272630,"speed":1,"start_troops":{"infantry_t1":20,"cavalry_t1":100},"troops":{"infantry_t1":20,"cavalry_t1":100},"opp_start_troops":[],"opp_troops":[],"troops_sentto_hospital":[],"resources":[],"rewards":[],"ctime":1422272310,"mtime":1422272310,"state":"marching","battle_log":[]}],"user_city":[{"city_troop":{"infantry_t1":87,"cavalry_t1":0,"ranged_t1":"100","artillery_close_t1":"100","artillery_distance_t1":"100"},"uid":45,"city_id":45}]}}}`
	p := new(Push)
	p.Body = []byte(body)
	for i := 0; i < b.N; i++ {
		p.Type()
	}
}

func BenchmarkDefer(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		defer func() {

		}()
	}
}

func BenchmarkSwitch(b *testing.B) {
	var x = 10
	for i := 0; i < b.N; i++ {
		switch x {
		case 1:
		case 2:
		case 4:
		case 10:
		default:
		}
	}
}

func BenchmarkAdd(b *testing.B) {
	b.ReportAllocs()
	var x int64
	for i := 0; i < b.N; i++ {
		x = x + 1
	}
}

func BenchmarkBitShift(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = 12 << 22
	}

}

func BenchmarkJobMarshal(b *testing.B) {
	b.ReportAllocs()
	job := Job{Uid: 534343}
	for i := 0; i < b.N; i++ {
		job.Marshal()
	}
	b.SetBytes(int64(len(job.Marshal())))
}

func BenchmarkMarchMarshal(b *testing.B) {
	b.ReportAllocs()
	march := March{Uid: 232323, MarchId: 23223232, State: "marching", X1: 12, Y1: 122, EndTime: time.Now()}
	for i := 0; i < b.N; i++ {
		march.Marshal()
	}
	b.SetBytes(int64(len(march.Marshal())))
}

func BenchmarkPveMarshal(b *testing.B) {
	b.ReportAllocs()
	pve := Pve{Uid: 3434343, MarchId: 343433434333, State: "marching", EndTime: time.Now()}
	for i := 0; i < b.N; i++ {
		pve.Marshal()
	}
	b.SetBytes(int64(len(pve.Marshal())))
}

func BenchmarkMutex(b *testing.B) {
	b.ReportAllocs()
	var mutex sync.Mutex
	for i := 0; i < b.N; i++ {
		mutex.Lock()
		mutex.Unlock()
	}
}

func BenchmarkBreakerSucceed(b *testing.B) {
	b.ReportAllocs()
	breaker := breaker.Consecutive{}
	for i := 0; i < b.N; i++ {
		breaker.Succeed()
	}
}

func BenchmarkPhpPayloadPartialDecode(b *testing.B) {
	b.ReportAllocs()
	payload := []byte(`{"ok":0,"msg":"0:Unknown event: . -- #0 \/sgn\/htdocs\/dev-dev\/dragon-server-code\/v2\/classes\/Services\/ActorService.php(20): Event\\EventEngine::fire(NULL, Array)\n#1 \/sgn\/htdocs\/dev-dev\/dragon-server-code\/v2\/classes\/System\/Application.php(160): Services\\ActorService->play(Array)\n#2 \/sgn\/htdocs\/dev-dev\/dragon-server-code\/docroot\/api\/index.php(12): System\\Application->execute()\n#3 {main}"}`)
	var (
		objmap map[string]*json.RawMessage
		ok     int
	)
	for i := 0; i < b.N; i++ {
		json.Unmarshal(payload, &objmap)
		json.Unmarshal(*objmap["ok"], &ok)
	}
}

func BenchmarkPhpPayloadFullDecode(b *testing.B) {
	b.ReportAllocs()
	payload := []byte(`{"ok":0,"msg":"0:Unknown event: . -- #0 \/sgn\/htdocs\/dev-dev\/dragon-server-code\/v2\/classes\/Services\/ActorService.php(20): Event\\EventEngine::fire(NULL, Array)\n#1 \/sgn\/htdocs\/dev-dev\/dragon-server-code\/v2\/classes\/System\/Application.php(160): Services\\ActorService->play(Array)\n#2 \/sgn\/htdocs\/dev-dev\/dragon-server-code\/docroot\/api\/index.php(12): System\\Application->execute()\n#3 {main}"}`)
	var decoded map[string]interface{}
	for i := 0; i < b.N; i++ {
		json.Unmarshal(payload, &decoded)
	}
}

// test json omit
func TestJobEncode(t *testing.T) {
	job := Job{Uid: 534343}
	body, _ := json.Marshal(job)
	assert.Equal(t, `{"uid":534343}`, string(body))
}
