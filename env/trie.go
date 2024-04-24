//go:build !nonwasmenv

package env

/*
	Trie: Interface that provides trie related functionality
*/

//go:wasmimport env ext_trie_blake2_256_ordered_root_version_2
func ExtTrieBlake2256OrderedRootVersion2(input int64, version int32) int32

// TODO
// ext_trie_blake2_256_root_version_1
// ext_trie_blake2_256_root_version_2
// ext_trie_blake2_256_ordered_root_version_1
// ext_trie_blake2_256_verify_proof_version_1
// ext_trie_blake2_256_verify_proof_version_2
// ext_trie_keccak_256_root_version_1
// ext_trie_keccak_256_root_version_2
// ext_trie_keccak_256_ordered_root_version_1
// ext_trie_keccak_256_ordered_root_version_2
// ext_trie_keccak_256_verify_proof_version_1
// ext_trie_keccak_256_verify_proof_version_2
