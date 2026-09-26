package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/goforj/godump"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"
)

func TestQueryRow(t *testing.T) {
	fmt.Println("addr", os.Getenv("SERVER_ADDR"))
	var str string
	err := QueryRow(context.Background(), "SELECT version();", nil, &str)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(str)
}

func TestSearchBots(t *testing.T) {

	out, err := SearchBots([]string{"search"})
	assert.NoError(t, err)
	godump.Dump(out)
}
