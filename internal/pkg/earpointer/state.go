// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package earpoint introduces a pointer analysis based on unifying
// equivalent abstract references (EAR), which implements Steensgaard's algorithm.
// Extensions include adding context sensitivity and field sensitivity.
// Abstract references defines the data structure to represent abstract heap.
package earpointer

import (
	"log"
	"sort"
	"strings"
)

// ParentMap maps a reference to its representative (i.e. the parent
// in the union-find tree).
type ParentMap map[Reference]Reference

// PartitionInfo maintains information at the partition level (which
// consists of a set of references). It maintains the size of the
// partition (used for weighted join in union-find) and a ReferenceFieldMap,
// which is a mapping from values F (instance fields, array contents etc) to
// references R.
type PartitionInfo struct {
	numMembers uint              // number of members
	fieldRefs  ReferenceFieldMap // field mapping
}

// PartitionInfoMap maps a reference to its partition.
type PartitionInfoMap map[Reference]PartitionInfo

// State is Equivalent Abstract References (EAR) state, which maintains
// an abstract heap information. Public APIs are provided to access the heap
// such as the partition information and the field maps.
type State struct {
	// Map from a ref (a partition representative) to information about
	// this partition.
	partitions PartitionInfoMap
	// Map from a ref to its parent abstract ref.
	parents ParentMap

	// Set to true after all mutation operations have completed: only lookups will
	// be performed from here on.
	finalized bool

	// Mapping from a partition representative to its members. This map is
	// initialize in Finalize method, which guarantees to avoid duplication.
	// Users are accessing this information via GetPartitionMembers method,
	// which does not allow any modification.
	members map[Reference][]Reference

	// Map from each field reference to its parent references. For example, record
	// r->o for "o.x = r". This map is for accelerating external queries on
	// fields, and is not required by the EAR analysis itself.
	// It is constructed separately using the "ConstructFieldParentMap()"
	// at the final phase.
	revFields map[Reference][]Reference
}

// NewState creates an empty abstract state.
func NewState() *State {
	return &State{
		partitions: make(PartitionInfoMap),
		parents:    make(ParentMap),
		finalized:  false,
		members:    make(map[Reference][]Reference),
		revFields:  make(map[Reference][]Reference),
	}
}

// HasRef returns true if "aref" is a reference present in the state.
func (state *State) HasRef(aref Reference) bool {
	_, ok := state.parents[aref]
	return ok
}

// GetAllRefs get all references.
func (state *State) AllRefs() ReferenceSet {
	refs := make(ReferenceSet)
	for k := range state.parents {
		refs[k] = true
	}
	return refs
}

// AllPartitionReps gets all references that are partition representatives.
func (state *State) AllPartitionReps() ReferenceSet {
	reps := make(ReferenceSet)
	for k, v := range state.parents {
		if k == v {
			reps[k] = true
		}
	}
	return reps
}

// GetPartitionRep gets the partition representative of reference "aref"
// ("aref" must belong to this state).
func (state *State) GetPartitionRep(aref Reference) Reference {
	failureCallback := func(aref Reference) Reference {
		log.Fatalf("GetPartitionRep: Reference [%s] not found in state", aref)
		return nil
	}
	return state.lookupPartitionRep(aref, failureCallback)
}

// GetPartitionFieldMap gets the partition-level field map for partition
// representative "aref". The result maps a field (ConstTermHandle type)
// to a reference that is a partition representative.
// Note "aref" must be a partition representative.
func (state *State) GetPartitionFieldMap(aref Reference) ReferenceFieldMap {
	return state.getPartitionInfo(aref).fieldRefs
}

// Pretty printer
func (state *State) String() string {
	arefs := state.AllRefs()
	members := make(map[Reference][]Reference)
	for aref := range arefs {
		rep := state.GetPartitionRep(aref)
		members[rep] = append(members[rep], aref)
	}

	var pstrs []string
	for p, s := range members {
		var mstrs []string
		for _, v := range s {
			mstrs = append(mstrs, v.String())
		}
		sort.Strings(mstrs)
		pstr := "{" + strings.Join(mstrs, ",") + "}: " +
			state.GetPartitionFieldMap(p).String()
		pstrs = append(pstrs, pstr)
	}
	sort.Strings(pstrs)
	return strings.Join(pstrs, ", ")
}

// Pretty prints a field map.
func (fmap ReferenceFieldMap) String() string {
	if v, ok := fmap[getDirectPointToField()]; ok {
		return "--> " + v.String()
	}
	var fstrs []string
	for k, v := range fmap {
		fstrs = append(fstrs, k.Name+"->"+v.String())
	}
	sort.Strings(fstrs)
	return "[" + strings.Join(fstrs, ", ") + "]"
}

// Various state mutation operation

// Insert inserts reference "aref" to the state and returns the current
// partition representative of "aref".
func (state *State) Insert(aref Reference) Reference {
	if state.finalized {
		log.Fatal("No mutations allowed after Finalize()")
	}
	// Lookup failure will create new entry in the parent table.
	failureCallback := func(ref Reference) Reference {
		state.parents[ref] = ref
		state.partitions[ref] = PartitionInfo{numMembers: 0, fieldRefs: make(ReferenceFieldMap)}
		return ref
	}
	return state.lookupPartitionRep(aref, failureCallback)
}

// Unify unifies the references "aref1" and "aref2"
func (state *State) Unify(aref1 Reference, aref2 Reference) {
	state.UnifyReps(state.GetPartitionRep(aref1), state.GetPartitionRep(aref2))
}

// UnifyReps unifies the references "aref1" and "aref2", assuming both are
// partition representative. This allows to avoid calling GetPartitionRep
// unnecesarilly. Using this function with non-representative
// references will raise an error.
func (state *State) UnifyReps(aref1 Reference, aref2 Reference) {
	if state.finalized {
		log.Fatal("No mutations allowed after Finalize()")
	}
	// If they are in the same partition already, we are done.
	if aref1 == aref2 {
		return
	}
	// Create local copies to use swap (this improves the readability).
	prep1, prep2 := aref1, aref2
	// Find partition-level info for "aref1" and "aref2"
	pinfo1 := state.getPartitionInfo(prep1)
	pinfo2 := state.getPartitionInfo(prep2)
	// swap so that "aref1" is always the smaller partition
	if pinfo1.numMembers > pinfo2.numMembers {
		prep1, prep2 = prep2, prep1
		pinfo1, pinfo2 = pinfo2, pinfo1
	}

	// Create state by having "prep1" point to "prep2" as parent and
	// by erasing "prep2" as a partition rep. We then call MergeFieldMap()
	// to merge field maps (which can trigger further unification). This process
	// will converge since the number of partitions is guaranteed to decrease at
	// every unification and is lower-bounded by 1.

	// Update parent pointer and partition info map for aref1 while retaining
	// its field map in fmap1 and number of members in n1
	state.parents[prep1] = prep2
	fmap1 := pinfo1.fieldRefs
	pinfo1.fieldRefs = make(ReferenceFieldMap)
	n1 := pinfo1.numMembers
	delete(state.partitions, prep1)

	// Update partition info for aref2 and merge the field map fmap1 in-place.
	pinfo2.numMembers += n1
	state.mergeFieldMap(fmap1, pinfo2.fieldRefs)
}

// Finalize indicates that no more mutation operations will be performed.
// After this call, internal data structures are optimized for lookups only,
// and no mutations are allowed.
func (state *State) Finalize() {
	// Construct the map from field references to their parents.
	state.constructFieldParentMap()

	state.finalized = true
	// Construct the members. members maps partitions representatives to the
	// members (References) of partitions.
	for aref, p := range state.parents {
		if aref == p {
			// If aref is equal to its parent, then aref is partition
			// representative. No need to call GetPartitionRep.
			state.members[aref] = append(state.members[aref], aref)
		} else {
			rep := state.GetPartitionRep(aref)
			// Otherwise, get aref's partition representative and update members.
			state.members[rep] = append(state.members[rep], aref)
		}
	}
	// Normalize fieldRefs table in a way that it only contains partition
	// representatives.
	for _, pinfo := range state.partitions {
		for fd, entry := range pinfo.fieldRefs {
			pinfo.fieldRefs[fd] = state.GetPartitionRep(entry)
		}
	}
}

// API calls that can be called after Finalize().

// GetPartitionSize returns the number of representatives
// (i.e., the size of the partition).
func (state *State) GetPartitionSize() int {
	if state.finalized {
		log.Fatal("No mutations allowed after Finalize()")
	}
	return len(state.members)
}

// GetPartitionMembersForRep gets the list of members in the partition
// for which "prep" is the representative.
// Precondition: "prep" must be a partition representative. Available only
// when after Finalize() is called.
func (state *State) GetPartitionMembersForRep(rep Reference) []Reference {
	if !state.finalized {
		log.Fatal("Should be called only after Finalize")
	}
	members, ok := state.members[rep]
	if !ok {
		log.Fatalf("Passed in reference [%s] is not a partition rep", rep)
	}
	return members
}

// GetPartitionMembers gets the list of members in the partition to which "aref" belongs.
// Precondition: "aref" must be a reference present in the state.
// Available only when after Finalize() is called.
func (state *State) GetPartitionMembers(aref Reference) []Reference {
	return state.GetPartitionMembersForRep(state.GetPartitionRep(aref))
}

// GetFieldParents returns the parent of a field reference. Return nil if ref has no parent.
// For example, return o for r if "o[x -> r]" is in state. This shall be
// called only after the state has been finalized.
func (state *State) GetFieldParents(ref Reference) []Reference {
	if !state.finalized {
		log.Fatal("Should be called only after Finalize")
	}
	return state.revFields[ref]
}

// GetValueReferenceOrNil return the reference holding the value *r of an address r.
// Return nil if such a value reference doesn't exists.
func (state *State) GetValueReferenceOrNil(addr Reference) Reference {
	// In the heap, r --> *r is implemented as r[PointToMarker] = *r.
	fmap := state.GetPartitionFieldMap(addr)
	if v, ok := fmap[getDirectPointToField()]; ok {
		return v
	}
	return nil
}

// mergeFieldMap sets the partition info for "rep" as a merge of the field
// maps in "oldMap" and "newMap" (the last is an out-parameter to save
// on memory copies).
func (state *State) mergeFieldMap(oldMap ReferenceFieldMap,
	newMap ReferenceFieldMap) {
	// Merge field maps of "oldMap" and update "newMap" suitably. For any
	// common fields, we unify the pointers, else we copy over the field mapping
	// as such.  We first compute the list of references to unify, update the
	// field map, and then perform unification.
	toUnify := make(map[Reference]Reference)
	// Merging old fmap to a new fmap. We iterate over oldMap and insert all elements
	// to newMap. If a field already exists in the newMap, we unify the corresponding heap
	// partitions. This takes O(mlog(n)) times, where 'n' is the size of newMap.
	for k, v := range oldMap {
		w, ok := newMap[k]
		if !ok {
			newMap[k] = v
		} else {
			toUnify[v] = w
		}
	}
	// Unify all necessary field partitions.
	for k, v := range toUnify {
		state.Unify(k, v)
	}
}

// Helper to obtain the partition info for "aref" (which should be a partition
// representative)
func (state *State) getPartitionInfo(aref Reference) *PartitionInfo {
	v, ok := state.partitions[aref]
	if !ok {
		log.Fatalf("Reference [%s] is not a partition rep", aref)
	}
	return &v
}

// Helper to find a partition representative. It calls onfailure callback
// to delegate the action on lookup failure.
func (state *State) lookupPartitionRep(
	aref Reference,
	onfailure func(abs Reference) Reference) Reference {
	ref := aref
	rep, ok := state.parents[ref]
	if !ok {
		res := onfailure(ref)
		return res
	}

	// If we found the partition rep, return right away.
	if rep == ref {
		return rep
	}
	// Else recurse. We use crashing callback here, because when recursing
	// we expect the recursion argument to be always in the abstract state.
	failureCallback := func(aref Reference) Reference {
		log.Fatal("lookupPartitionRep: Reference not found in state")
		return nil
	}
	prep := state.lookupPartitionRep(rep, failureCallback)
	// Perform path compression before returning to caller.
	state.parents[aref] = prep
	return prep
}

// constructFieldParentMap constructs (from scratch) a map from each field reference
// to its parent reference. For example, record r->o for "o.x = r".
func (state *State) constructFieldParentMap() {
	if state.finalized {
		log.Fatal("No mutations allowed after Finalize()")
	}
	reps := state.AllPartitionReps()
	for rep := range reps {
		if fmap := state.GetPartitionFieldMap(rep); fmap != nil {
			for _, fref := range fmap {
				state.revFields[fref] = append(state.revFields[fref], rep)
			}
		}
	}
}