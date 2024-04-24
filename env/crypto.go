//go:build !nonwasmenv

package env

/*
	Crypto: Interfaces for working with crypto related types from within the runtime.
*/

//go:wasmimport env ext_crypto_ed25519_generate_version_1
func ExtCryptoEd25519GenerateVersion1(key_type_id int32, seed int64) int32

//go:wasmimport env ext_crypto_ed25519_verify_version_1
func ExtCryptoEd25519VerifyVersion1(sig int32, msg int64, key int32) int32

//go:wasmimport env ext_crypto_secp256k1_ecdsa_recover_version_2
func ExtCryptoSecp256k1EcdsaRecoverVersion2(sig int32, msg int32) int64

//go:wasmimport env ext_crypto_secp256k1_ecdsa_recover_compressed_version_2
func ExtCryptoSecp256k1EcdsaRecoverCompressedVersion2(sig int32, msg int32) int64

//go:wasmimport env ext_crypto_sr25519_generate_version_1
func ExtCryptoSr25519GenerateVersion1(key_type_id int32, seed int64) int32

//go:wasmimport env ext_crypto_sr25519_public_keys_version_1
func ExtCryptoSr25519PublicKeysVersion1(key_type_id int32) int64

//go:wasmimport env ext_crypto_sr25519_sign_version_1
func ExtCryptoSr25519SignVersion1(key_type_id int32, key int32, msg int64) int64

//go:wasmimport env ext_crypto_sr25519_verify_version_2
func ExtCryptoSr25519VerifyVersion2(sig int32, msg int64, key int32) int32

// TODO
// ext_crypto_ed25519_public_keys_version_1
// ext_crypto_ed25519_sign_version_1
// ext_crypto_secp256k1_ecdsa_recover_version_1
// ext_crypto_ecdsa_verify_version_2
// ext_crypto_secp256k1_ecdsa_recover_compressed_version_1
// ext_crypto_sr25519_verify_version_1
// ext_crypto_start_batch_verify_version_1
// ext_crypto_finish_batch_verify_version_1
// ext_crypto_ecdsa_generate_version_1
// ext_crypto_sr25519_batch_verify_version_1
// ext_crypto_ecdsa_public_key_version_1
// ext_crypto_ecdsa_sign_version_1
// ext_crypto_ecdsa_sign_prehashed_version_1
// ext_crypto_ecdsa_verify_version_1
// ext_crypto_ecdsa_verify_prehashed_version_1
// ext_crypto_ecdsa_batch_verify_version_1
