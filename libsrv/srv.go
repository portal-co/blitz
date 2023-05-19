package libsrv

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strings"

	"blitz.build/types"
)

func Serve(x func(map[string]string) (map[string]types.Pathed, error)) error {
	h := http.NewServeMux()
	h.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		os.Remove(os.Args[2])
	})
	h.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		n := r.URL.Query()["input"][0]
		m := map[string]string{}
		for _, x := range strings.Split(n, ";") {
			a := strings.SplitN(x, "=", 2)
			m[a[0]] = a[1]
		}
		i, err := x(m)
		if err == nil {
			json.NewEncoder(w).Encode(i)
		}
	})

	conn, err := net.Listen("unix", os.Args[2])
	if err != nil {
		return err
	}

	defer conn.Close()
	return http.Serve(conn, h)
}
