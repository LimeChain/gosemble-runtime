//go:build !nonwasmenv

package env

// TODO
/*
	Offchain:
*/

// ext_offchain_local_storage_clear_version_1
// ext_offchain_is_validator_version_1
// ext_offchain_local_storage_compare_and_set_version_1
// ext_offchain_local_storage_get_version_1
// ext_offchain_local_storage_set_version_1
// ext_offchain_network_state_version_1
// ext_offchain_random_seed_version_1

//go:wasmimport env ext_offchain_submit_transaction_version_1
func ExtOffchainSubmitTransactionVersion1(data int64) int64

// ext_offchain_timestamp_version_1
// ext_offchain_sleep_until_version_1
// ext_offchain_http_request_start_version_1
// ext_offchain_http_request_add_header_version_1
// ext_offchain_http_request_write_body_version_1
// ext_offchain_http_response_wait_version_1
// ext_offchain_http_response_headers_version_1
// ext_offchain_http_response_read_body_version_1
