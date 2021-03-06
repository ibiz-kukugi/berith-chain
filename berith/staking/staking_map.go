package staking

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"sort"

	"github.com/BerithFoundation/berith-chain/core/state"

	"github.com/BerithFoundation/berith-chain/common"
	"github.com/BerithFoundation/berith-chain/rlp"
)

//[BERITH]
//StakingMap StakingList 인터페이스를 맵형태로 구현한 구조체
type StakingMap struct {
	storage    map[common.Address]stkInfo
	sortedList []common.Address
	miners     map[common.Address]bool
	table      map[common.Address]VoteResult
}

//[BERITH]
//토큰을 예치한 계정의 정보를 나타내는 구조체
type stkInfo struct {
	StkAddress     common.Address `json:"address"`     //토큰을 예치한 계정
	StkValue       *big.Int       `json:"value"`       //예치한 토큰의 수량
	StkBlockNumber *big.Int       `json:"blocknumber"` //토큰을 예치한 시점의 블록번호

}

//[BERITH]
//stkInfo 구조체의 Getter 메서드
func (s stkInfo) Address() common.Address { return s.StkAddress }
func (s stkInfo) Value() *big.Int         { return s.StkValue }
func (s stkInfo) BlockNumber() *big.Int   { return s.StkBlockNumber }

//[BERITH]
//Len 목록의 길이를 반환하는 메서드
func (list *StakingMap) Len() int {
	return len(list.sortedList)
}

//[BERITH]
// 특정 계정이 블록을 생성할 때의 난이도와, 순위를 반환하는 메서드
func (list *StakingMap) GetDifficultyAndRank(addr common.Address, blockNumber uint64, states *state.StateDB) (*big.Int, int, bool) {
	reordered := false
	if len(list.table) <= 0 {
		reordered = true
		list.selectSigner(blockNumber, states)
	}
	if len(list.table) <= 0 {
		return big.NewInt(1234), 1, false
	}

	result, ok := list.table[addr]
	if !ok {
		result = VoteResult{
			Score: big.NewInt(0),
			Rank:  -1,
		}
	}
	return result.Score, result.Rank, reordered
}

//[BERITH]
//GetInfo 특정 계정의 "StakingInfo" 를 반환하는 메서드
func (list *StakingMap) GetInfo(address common.Address) (StakingInfo, error) {
	info, ok := list.storage[address]

	if !ok {
		return &stkInfo{
			StkAddress:     address,
			StkValue:       big.NewInt(0),
			StkBlockNumber: big.NewInt(0),
		}, nil
	}
	return &stkInfo{
		StkAddress:     address,
		StkValue:       info.Value(),
		StkBlockNumber: info.BlockNumber(),
	}, nil
}

//[BERITH]
//SetInfo 목록에 "StakingInfo" 를 등록하는 메서드
func (list *StakingMap) SetInfo(info StakingInfo) error {

	if info.Value().Cmp(big.NewInt(0)) < 1 {
		delete(list.storage, info.Address())
		return nil
	}

	list.storage[info.Address()] = stkInfo{
		StkAddress:     info.Address(),
		StkValue:       info.Value(),
		StkBlockNumber: info.BlockNumber(),
	}
	return nil
}

//[BERITH]
//ToArray StakingMap의 내용을 배열형태로 반환하는 메서드
func (list *StakingMap) ToArray() []common.Address {
	return list.sortedList
}

//[BERITH]
//Delete 특정 계정의 "StakingInfo" 를 삭제하는 메서드
func (list *StakingMap) Delete(address common.Address) error {
	if _, ok := list.storage[address]; ok {
		delete(list.storage, address)
	}
	return nil
}

//[BERITH]
//Print StakingMap에 대한 로그를 콘솔화면에 출력하는 메서드
func (list *StakingMap) Print() {
	fmt.Println("==== Staking List ====")
	for k, v := range list.storage {
		fmt.Println("** [key : ", k.Hex(), " | value : ", v.Value().String(), "| blockNumber : ", v.BlockNumber().String(), "]")
	}
	fmt.Println("==== sortedList ====")
	for _, v := range list.sortedList {
		fmt.Println(v.Hex())
	}
	// fmt.Println("====== TABLE ======")
	// for k, v := range list.table {
	// 	fmt.Println("[", k.Hex(), ",", v.Rank, ",", v.Score.String(), "]")
	// }
}

//[BERITH]
//EncodeRLP StakingMap 구조체를 RLP 인코딩 하기 위한 메서드
func (list *StakingMap) EncodeRLP(w io.Writer) error {

	var byteArr [4][]byte

	byteArr[0], _ = json.Marshal(list.storage)
	byteArr[1], _ = json.Marshal(list.miners)
	byteArr[2], _ = json.Marshal(list.table)
	byteArr[3], _ = json.Marshal(list.sortedList)
	//rlpVal[1], _ = json.Marshal(list.sortedList)
	return rlp.Encode(w, byteArr)
}

//[BERITH]
//Sort 목록을 정렬하기 위한 메서드
func (list *StakingMap) Sort() {

	if len(list.sortedList) > 0 {
		return
	}

	kv := make(infoForSort, 0)
	for _, v := range list.storage {
		if v.Value().Cmp(big.NewInt(0)) > 0 {
			kv = append(kv, v)
		}
	}
	sort.Sort(&kv)

	sortedList := make([]common.Address, 0)

	for _, info := range kv {

		sortedList = append(sortedList, info.Address())
	}

	list.sortedList = sortedList
}

func (list *StakingMap) ClearTable() {
	list.sortedList = make([]common.Address, 0)
	list.table = make(map[common.Address]VoteResult)
}

//[BERITH]
//selectSigner 전체 목록중에서 블록을 생성할 유저를 선별한 결과를 반환하는 메서드
func (list *StakingMap) selectSigner(blockNumber uint64, states *state.StateDB) {

	if len(list.sortedList) <= 0 {
		list.Sort()
	}

	if len(list.sortedList) <= 0 {
		return
	}

	cs := NewCandidates()

	for _, addr := range list.sortedList {
		info := list.storage[addr]
		cs.Add(Candidate{info.Address(), states.GetPoint(info.Address()).Uint64(), 0})
		cs.ts += new(big.Int).Div(states.GetStakeBalance(info.Address()), big.NewInt(1e+18)).Uint64()
	}

	list.table = *cs.BlockCreator(blockNumber)

	//for key, value := range list.table {
	//	fmt.Println("ADDRESS :: "+key.String(), "DIFF :: "+value.String())
	//}

}

//[BERITH]
//GetJoinRatio 특정계정이 블록을 생성할 확률을 반환하는 메서드
func (list *StakingMap) GetJoinRatio(address common.Address, blockNumber uint64, states *state.StateDB) float64 {
	cs := NewCandidates()

	for _, addr := range list.sortedList {
		info := list.storage[addr]
		cs.Add(Candidate{info.Address(), states.GetPoint(info.Address()).Uint64(), 0})
	}
	roi := cs.getJoinRatio(address)
	return roi
}

//[BERITH]
//infoForSort 목록을 정렬할 때, 사용되는 배열
type infoForSort []stkInfo

func (info infoForSort) Len() int { return len(info) }
func (info infoForSort) Less(i, j int) bool {
	if info[i].Value().Cmp(info[j].Value()) == 0 {
		if info[i].BlockNumber().Cmp(info[j].BlockNumber()) == 0 {
			return bytes.Compare(info[i].Address().Bytes(), info[j].Address().Bytes()) > 0
		}
		return info[i].BlockNumber().Cmp(info[j].BlockNumber()) < 0
	}
	return info[i].Value().Cmp(info[j].Value()) > 0
}
func (info infoForSort) Swap(i, j int) { info[i], info[j] = info[j], info[i] }

//[BERITH]
//Copy 구조체를 복사하여 반환하는 메서드
func (list *StakingMap) Copy() StakingList {
	return &StakingMap{
		storage:    list.storage,
		sortedList: list.sortedList,
		miners:     list.miners,
		table:      list.table,
	}
}

//[BERITH]
//Encode StakingMap 구조체를 RLP 인코딩한 바이트값을 반환하는 메서드
func (list *StakingMap) Encode() ([]byte, error) {
	return rlp.EncodeToBytes(list)
}

//[BERITH]
//Decode 바이트 배열을 디코딩하여 StakingMap 구조체를 반환하는 메서드
func (list *StakingMap) Decode(rlpData []byte) (StakingList, error) {
	return Decode(rlpData)
}

//[BERITH]
//Encode StakingMap 구조체를 입력받아 RLP 인코딩한 바이트값을 반환하는 함수
func Encode(stakingList StakingList) ([]byte, error) {
	return rlp.EncodeToBytes(stakingList)
}

//[BERITH]
//Decode 바이트 배열을 디코딩하여 StakingMap 구조체를 반환하는 함수
func Decode(rlpData []byte) (StakingList, error) {
	var byteArr [4][]byte
	if err := rlp.DecodeBytes(rlpData, &byteArr); err != nil {
		return nil, err
	}

	result := &StakingMap{
		storage:    make(map[common.Address]stkInfo),
		sortedList: make([]common.Address, 0),
		miners:     make(map[common.Address]bool),
		table:      make(map[common.Address]VoteResult),
	}
	if err := json.Unmarshal(byteArr[0], &result.storage); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(byteArr[1], &result.miners); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(byteArr[2], &result.table); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(byteArr[3], &result.sortedList); err != nil {
		return nil, err
	}

	return result, nil

}

//[BERITH]
//New 새로운 StakinMap 구조체를 생성하는 함수
func New() StakingList {
	return &StakingMap{
		storage:    make(map[common.Address]stkInfo),
		sortedList: make([]common.Address, 0),
		miners:     make(map[common.Address]bool),
		table:      make(map[common.Address]VoteResult),
	}
}
