package manual

/*
#cgo CFLAGS: -I/usr/include/
#cgo LDFLAGS: -Wl,-Bstatic -L/usr/lib/ -ltokenizers_c -lm -lpthread -Wl,-Bdynamic
#include "tokenizers_c.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"unsafe"
)

type Tokenizer struct {
	handle C.TokenizerHandle
}

func NewFromJSON(jsonStr string) (*Tokenizer, error) {
	cJSON := C.CString(jsonStr)
	defer C.free(unsafe.Pointer(cJSON))

	handle := C.tokenizers_new_from_str(cJSON, C.size_t(len(jsonStr)))
	if handle == nil {
		return nil, errors.New("failed to create tokenizer from JSON")
	}

	return &Tokenizer{handle: handle}, nil
}

func NewByteLevelBPEFromStr(vocab, merges, addedTokens string) (*Tokenizer, error) {
	cVocab := C.CString(vocab)
	defer C.free(unsafe.Pointer(cVocab))
	cMerges := C.CString(merges)
	defer C.free(unsafe.Pointer(cMerges))
	cAdded := C.CString(addedTokens)
	defer C.free(unsafe.Pointer(cAdded))

	handle := C.byte_level_bpe_tokenizers_new_from_str(
		cVocab, C.size_t(len(vocab)),
		cMerges, C.size_t(len(merges)),
		cAdded, C.size_t(len(addedTokens)),
	)
	if handle == nil {
		return nil, errors.New("failed to create byte-level BPE tokenizer")
	}
	return &Tokenizer{handle: handle}, nil
}

func (t *Tokenizer) Free() {
	if t.handle != nil {
		C.tokenizers_free(t.handle)
		t.handle = nil
	}
}

func (t *Tokenizer) Encode(text string, addSpecialToken bool) ([]int32, error) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	var result C.TokenizerEncodeResult
	var addFlag C.int = 0
	if addSpecialToken {
		addFlag = 1
	}

	C.tokenizers_encode(t.handle, cText, C.size_t(len(text)), addFlag, &result)
	if result.token_ids == nil {
		return nil, errors.New("failed to encode text")
	}
	length := int(result.len)
	cTokens := (*[1 << 28]C.int)(unsafe.Pointer(result.token_ids))[:length:length]
	goTokens := make([]int32, length)
	for i, token := range cTokens {
		goTokens[i] = int32(token)
	}

	C.tokenizers_free_encode_results(&result, 1)
	return goTokens, nil
}

func (t *Tokenizer) EncodeBatch(texts []string, addSpecialToken bool) ([][]int32, error) {
	numSeqs := len(texts)
	if numSeqs == 0 {
		return nil, errors.New("no texts provided")
	}

	cTexts := make([]*C.char, numSeqs)
	cLens := make([]C.size_t, numSeqs)
	for i, text := range texts {
		cTexts[i] = C.CString(text)
		cLens[i] = C.size_t(len(text))
		defer C.free(unsafe.Pointer(cTexts[i]))
	}

	results := make([]C.TokenizerEncodeResult, numSeqs)

	C.tokenizers_encode_batch(
		t.handle,
		(**C.char)(unsafe.Pointer(&cTexts[0])),
		(*C.size_t)(unsafe.Pointer(&cLens[0])),
		C.size_t(numSeqs),
		C.int(boolToInt(addSpecialToken)),
		(*C.TokenizerEncodeResult)(unsafe.Pointer(&results[0])),
	)

	batch := make([][]int32, numSeqs)
	for i := 0; i < numSeqs; i++ {
		length := int(results[i].len)
		cTokens := (*[1 << 28]C.int)(unsafe.Pointer(results[i].token_ids))[:length:length]
		goTokens := make([]int32, length)
		for j, token := range cTokens {
			goTokens[j] = int32(token)
		}
		batch[i] = goTokens
	}

	C.tokenizers_free_encode_results((*C.TokenizerEncodeResult)(unsafe.Pointer(&results[0])), C.size_t(numSeqs))
	return batch, nil
}

func (t *Tokenizer) Decode(tokenIDs []uint32, skipSpecialToken bool) (string, error) {
	if len(tokenIDs) == 0 {
		return "", errors.New("no token IDs provided")
	}
	cTokens := make([]C.uint32_t, len(tokenIDs))
	for i, id := range tokenIDs {
		cTokens[i] = C.uint32_t(id)
	}

	var skipFlag C.int = 0
	if skipSpecialToken {
		skipFlag = 1
	}
	C.tokenizers_decode(t.handle, (*C.uint32_t)(unsafe.Pointer(&cTokens[0])), C.size_t(len(cTokens)), skipFlag)

	var cStr *C.char
	var strLen C.size_t
	C.tokenizers_get_decode_str(t.handle, (**C.char)(unsafe.Pointer(&cStr)), &strLen)
	if cStr == nil {
		return "", errors.New("failed to decode tokens")
	}
	decoded := C.GoStringN(cStr, C.int(strLen))
	return decoded, nil
}

func (t *Tokenizer) GetVocabSize() (int, error) {
	var size C.size_t
	C.tokenizers_get_vocab_size(t.handle, &size)
	return int(size), nil
}

func (t *Tokenizer) IdToToken(id uint32) (string, error) {
	var cStr *C.char
	var strLen C.size_t
	C.tokenizers_id_to_token(t.handle, C.uint32_t(id), (**C.char)(unsafe.Pointer(&cStr)), &strLen)
	if cStr == nil {
		return "", errors.New("failed to convert id to token")
	}
	token := C.GoStringN(cStr, C.int(strLen))
	return token, nil
}

func (t *Tokenizer) TokenToId(token string) (int32, error) {
	cToken := C.CString(token)
	defer C.free(unsafe.Pointer(cToken))
	var id C.int32_t
	C.tokenizers_token_to_id(t.handle, cToken, C.size_t(len(token)), &id)
	if id == -1 {
		return int32(id), errors.New("token not found in vocabulary")
	}
	return int32(id), nil
}

// Helper to convert bool to int.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
