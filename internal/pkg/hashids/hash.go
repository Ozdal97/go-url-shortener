package hashids

import (
	"crypto/rand"
	"math/big"

	"github.com/speps/go-hashids/v2"
)

type Encoder struct {
	h *hashids.HashID
}

func New(salt string, minLen int) (*Encoder, error) {
	d := hashids.NewData()
	d.Salt = salt
	d.MinLength = minLen
	d.Alphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	h, err := hashids.NewWithData(d)
	if err != nil {
		return nil, err
	}
	return &Encoder{h: h}, nil
}

func (e *Encoder) Encode(n int64) (string, error) {
	return e.h.EncodeInt64([]int64{n})
}

// Random küçük bir tamsayı üretip kodlar. Veritabanı id'si olmadan
// link oluşturmak gerektiğinde kullanılır.
func (e *Encoder) Random() (string, error) {
	max := big.NewInt(1 << 40)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return e.Encode(n.Int64())
}
