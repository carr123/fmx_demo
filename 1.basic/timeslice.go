package timeslice

import (
	"container/list"
	"sort"
	"sync"
	"time"
)

type TimeSliceT struct {
	Begin time.Time //开始时间
	End   time.Time //结束时间
	Tag1  string    //附加字段1
	Tag2  string    //附加字段2
}

type TimeSliceMgr struct {
	vec    *list.List
	locker sync.Mutex
}

func NewTimeSlice(begin time.Time, end time.Time, tag1 string, tag2 string) TimeSliceT {
	return TimeSliceT{
		Begin: begin,
		End:   end,
		Tag1:  tag1,
		Tag2:  tag2,
	}
}

func NewTimeSliceMgr(arr ...TimeSliceT) *TimeSliceMgr {
	obj := &TimeSliceMgr{
		vec: list.New(),
	}

	for _, item := range arr {
		obj.vec.PushBack(item)
	}

	return obj
}

func (t *TimeSliceMgr) InsertTimeSlices(arr ...TimeSliceT) bool {
	if t.vec.Len() == 0 {
		for _, item := range arr {
			t.vec.PushBack(item)
		}
		return true
	}

	newElements := make([]*list.Element, 0, len(arr))

	var begin *list.Element = t.vec.Front()
	var bAdded bool
	for _, item := range arr {
		bAdded = false
		for ite := begin; ite != nil; ite = ite.Next() {
			if ite.Prev() == nil && ite.Value.(TimeSliceT).Begin.Sub(item.End) >= 0 {
				begin = t.vec.InsertBefore(item, ite)
				newElements = append(newElements, begin)
				bAdded = true
				break
			}

			if ite.Next() == nil && item.Begin.Sub(ite.Value.(TimeSliceT).End) >= 0 {
				begin = t.vec.InsertAfter(item, ite)
				newElements = append(newElements, begin)
				bAdded = true
				break
			}

			nextel := ite.Next()
			if nextel != nil &&
				nextel.Value.(TimeSliceT).Begin.Sub(item.End) >= 0 &&
				item.Begin.Sub(ite.Value.(TimeSliceT).End) >= 0 {
				begin = t.vec.InsertAfter(item, ite)
				newElements = append(newElements, begin)
				bAdded = true
				break
			}
		}

		if !bAdded {
			break
		}
	}

	if bAdded {
		return true
	}

	for _, e := range newElements {
		t.vec.Remove(e)
	}

	return false
}

func (t *TimeSliceMgr) CanInsertTimeSlices(arr ...TimeSliceT) bool {
	if t.vec.Len() == 0 || len(arr) == 0 {
		return true
	}

	var begin *list.Element = t.vec.Front()
	var bAdded bool
	for _, item := range arr {
		bAdded = false
		for ite := begin; ite != nil; ite = ite.Next() {

			if ite.Prev() == nil && ite.Value.(TimeSliceT).Begin.Sub(item.End) >= 0 {
				bAdded = true
				break
			}

			if ite.Next() == nil && item.Begin.Sub(ite.Value.(TimeSliceT).End) >= 0 {
				begin = ite
				bAdded = true
				break
			}

			nextel := ite.Next()
			if nextel != nil &&
				nextel.Value.(TimeSliceT).Begin.Sub(item.End) >= 0 &&
				item.Begin.Sub(ite.Value.(TimeSliceT).End) >= 0 {
				begin = ite
				bAdded = true
				break
			}
		}

		if !bAdded {
			return false
		}
	}

	return true
}

/*
func (t *TimeSliceMgr) GetSpareTime(tbegin time.Time, tend time.Time) time.Duration {
	if t.vec.Len() == 0 ||
		tend.Sub(tbegin) < 0 ||
		t.vec.Front().Value.(TimeSliceT).Begin.Sub(tend) >= 0 ||
		tbegin.Sub(t.vec.Back().Value.(TimeSliceT).End) >= 0 {
		return tend.Sub(tbegin)
	}

	var nFree time.Duration
	d1 := t.vec.Front().Value.(TimeSliceT).Begin.Sub(tbegin)
	if d1 >= 0 {
		nFree += d1
	}

	for ite := t.vec.Front(); ite != nil; ite = ite.Next() {
		if ite.Value.(TimeSliceT).End.Sub(tend) >= 0 {
			break
		}

		nextSlice := ite.Next()
		if nextSlice != nil {
			if tbegin.Sub(nextSlice.Value.(TimeSliceT).Begin) >= 0 {
				continue
			}

			t1 := t._timemax(ite.Value.(TimeSliceT).End, tbegin)
			t2 := t._timemin(nextSlice.Value.(TimeSliceT).Begin, tend)
			if t._istimesorted(t1, t2) {
				nFree += t2.Sub(t1)
			} else {
				break
			}
		} else {
			t1 := t._timemax(ite.Value.(TimeSliceT).End, tbegin)
			if t._istimesorted(t1, tend) {
				nFree += tend.Sub(t1)
			} else {
				break
			}
		}
	}

	return nFree
}
*/
func (t *TimeSliceMgr) GetSpareTime(tbegin time.Time, tend time.Time) time.Duration {
	if t.vec.Len() == 0 ||
		tend.Sub(tbegin) < 0 ||
		t.vec.Front().Value.(TimeSliceT).Begin.Sub(tend) >= 0 ||
		tbegin.Sub(t.vec.Back().Value.(TimeSliceT).End) >= 0 {
		return tend.Sub(tbegin)
	}

	//找到与目标时间片有时间交叉的
	tmSlices := make([]TimeSliceT, 0, t.vec.Len())
	for ite := t.vec.Front(); ite != nil; ite = ite.Next() {
		if tbegin.Sub(ite.Value.(TimeSliceT).End) >= 0 {
			continue
		}

		if ite.Value.(TimeSliceT).Begin.Sub(tend) >= 0 {
			break
		}

		tmSlices = append(tmSlices, ite.Value.(TimeSliceT))
	}

	var nTotal time.Duration = tend.Sub(tbegin)

	for _, item := range tmSlices {
		tm1 := t._timemax(item.Begin, tbegin)
		tm2 := t._timemin(item.End, tend)
		nTotal -= tm2.Sub(tm1)
	}

	if nTotal < 0 {
		return 0
	}

	return nTotal
}

func (t *TimeSliceMgr) _timemin(t1 time.Time, t2 time.Time) time.Time {
	if t1.Sub(t2) >= 0 {
		return t2
	} else {
		return t1
	}
}

func (t *TimeSliceMgr) _timemax(t1 time.Time, t2 time.Time) time.Time {
	if t1.Sub(t2) >= 0 {
		return t1
	} else {
		return t2
	}
}

func (t *TimeSliceMgr) _istimesorted(t1 time.Time, t2 time.Time) bool {
	if t2.Sub(t1) >= 0 {
		return true
	} else {
		return false
	}
}

func (t *TimeSliceMgr) String() string {
	var s string
	for ite := t.vec.Front(); ite != nil; ite = ite.Next() {
		t1 := ite.Value.(TimeSliceT).Begin.Format("2006-01-02 15:04:05")
		t2 := ite.Value.(TimeSliceT).End.Format("2006-01-02 15:04:05")
		s = s + t1 + " - " + t2 + "\n"
	}
	return s
}

func SortTimeSliceAsc(arr []TimeSliceT) {
	sort.Slice(arr, func(i, j int) bool {
		if arr[i].Begin.Before(arr[j].Begin) {
			return true
		} else {
			return false
		}
	})
}

func IsTimeSliceConflict(s1 *TimeSliceT, s2 *TimeSliceT) bool {
	duration1 := s1.End.Sub(s1.Begin)
	duration2 := s2.End.Sub(s2.Begin)

	dates := []time.Time{s1.Begin, s1.End, s2.Begin, s2.End}
	sort.Slice(dates, func(i int, j int) bool {
		if dates[j].Sub(dates[i]) >= 0 {
			return true
		} else {
			return false
		}
	})

	duration3 := dates[3].Sub(dates[0])
	if duration3 >= duration1+duration2 {
		return false
	}

	return true
}
