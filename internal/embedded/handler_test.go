package embedded_test

import (
	"encoding/json"
	"fmt"
	"github.com/a-clap/iot/internal/embedded"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewFromConfig(t *testing.T) {
	t.Run("test config parsing", func(t *testing.T) {
		c := embedded.ConfigHeater{}
		c.Pin.Chip = "chip"
		c.Pin.Line = 1

		buf, err := json.Marshal(c)
		require.Nil(t, err)
		fmt.Println(string(buf))

	})

}
