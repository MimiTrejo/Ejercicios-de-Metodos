package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

// ============================================================
// METODO 2: DESCENDENTE PREDICTIVO NO RECURSIVO (tabla LL1)
//
// Gramatica:
//   S → a S b
//   S → a b
//
// Tabla LL(1):
//   M[S][a] = a S b   (cuando el siguiente de 'a' es otra 'a')
//   M[S][a] = a b     (cuando el siguiente de 'a' es 'b')
//
// NOTA: esta gramatica no es estrictamente LL(1) con un solo
// token de lookahead desde S, por eso usamos lookahead=2
// (el token actual + el siguiente) solo en el no-terminal S.
// ============================================================

func lexer(entrada string) ([]rune, error) {
	var tokens []rune
	for _, ch := range entrada {
		if ch == ' ' {
			continue
		}
		if ch != 'a' && ch != 'b' {
			return nil, fmt.Errorf("caracter invalido '%c'", ch)
		}
		tokens = append(tokens, ch)
	}
	tokens = append(tokens, '$')
	return tokens, nil
}

var nonTerminals = map[string]bool{"S": true}

type StepRow struct {
	Pila    string
	Entrada string
	Accion  string
}

func analizarLL1(cadena string) ([]StepRow, bool) {
	cadena = strings.TrimSpace(cadena)
	tokens, err := lexer(cadena)
	if err != nil {
		return []StepRow{{Accion: "Error Lexico: " + err.Error()}}, false
	}

	// pila: tope es el ultimo elemento
	stack := []string{"$", "S"}
	ip := 0

	stackStr := func() string {
		s := make([]string, len(stack))
		for i, v := range stack {
			s[len(stack)-1-i] = v
		}
		return strings.Join(s, " ")
	}
	inputStr := func() string {
		var parts []string
		for _, t := range tokens[ip:] {
			parts = append(parts, string(t))
		}
		return strings.Join(parts, " ")
	}

	var rows []StepRow
	ok := true

	for {
		top := stack[len(stack)-1]
		cur := tokens[ip]

		// Condicion de aceptacion
		if top == "$" && cur == '$' {
			rows = append(rows, StepRow{stackStr(), inputStr(), "ACEPTAR"})
			break
		}
		if top == "$" {
			rows = append(rows, StepRow{stackStr(), inputStr(), "Error: pila vacia pero quedan tokens"})
			ok = false
			break
		}

		if !nonTerminals[top] {
			// terminal: debe concordar
			if len(top) == 1 && rune(top[0]) == cur {
				rows = append(rows, StepRow{stackStr(), inputStr(), fmt.Sprintf("concordancia '%c'", cur)})
				stack = stack[:len(stack)-1]
				ip++
			} else {
				rows = append(rows, StepRow{stackStr(), inputStr(), fmt.Sprintf("Error: se esperaba '%s' y llego '%c'", top, cur)})
				ok = false
				break
			}
		} else {
			// no-terminal S: consultar tabla con lookahead 2
			if cur != 'a' {
				rows = append(rows, StepRow{stackStr(), inputStr(), fmt.Sprintf("Error: M[S]['%c'] vacio", cur)})
				ok = false
				break
			}
			// ver siguiente token para decidir produccion
			siguiente := '$'
			if ip+1 < len(tokens) {
				siguiente = tokens[ip+1]
			}

			var prod []string
			var accion string
			if siguiente == 'a' {
				prod = []string{"a", "S", "b"}
				accion = "S → a S b"
			} else if siguiente == 'b' {
				prod = []string{"a", "b"}
				accion = "S → a b"
			} else {
				rows = append(rows, StepRow{stackStr(), inputStr(), fmt.Sprintf("Error: M[S]['%c'] vacio", cur)})
				ok = false
				break
			}

			rows = append(rows, StepRow{stackStr(), inputStr(), accion})
			stack = stack[:len(stack)-1]
			for i := len(prod) - 1; i >= 0; i-- {
				stack = append(stack, prod[i])
			}
		}
	}
	return rows, ok
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Metodo 2 - LL(1) No Recursivo</title>
</head>
<body>

<h1>Metodo 2: Descendente Predictivo No Recursivo (LL1)</h1>

<pre>
Gramatica:
  S → a S b
  S → a b

Tabla LL(1):
  M[S][a, sig=a] = S → a S b
  M[S][a, sig=b] = S → a b

Ejemplos validos: ab, aabb, aaabbb
</pre>

<form method="POST">
<label>Ingrese cadenas (una por linea):</label><br><br>
<textarea name="cadenas" rows="6" cols="30">{{.Input}}</textarea><br><br>
<button type="submit">Analizar</button>
</form>

{{if .Resultados}}
<h2>Resultados</h2>

{{range .Resultados}}
<p><b>{{.Cadena}}</b> → {{if .Valida}}VALIDA{{else}}INVALIDA{{end}}</p>
<table border="1" cellpadding="4" cellspacing="0">
  <tr>
    <th>Pila (tope→)</th>
    <th>Entrada</th>
    <th>Accion</th>
  </tr>
  {{range .Pasos}}
  <tr>
    <td>{{.Pila}}</td>
    <td>{{.Entrada}}</td>
    <td>{{.Accion}}</td>
  </tr>
  {{end}}
</table>
<br>
{{end}}

<p>Total validas: {{.Validas}} / {{len .Resultados}}</p>
{{end}}

</body>
</html>
`

type Resultado struct {
	Cadena string
	Pasos  []StepRow
	Valida bool
}

func main() {
	type PageData struct {
		Input      string
		Resultados []Resultado
		Validas    int
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.New("web").Parse(htmlTemplate))
		data := PageData{}

		if r.Method == http.MethodPost {
			input := r.FormValue("cadenas")
			data.Input = input
			for _, linea := range strings.Split(input, "\n") {
				linea = strings.TrimSpace(linea)
				if linea == "" {
					continue
				}
				pasos, valida := analizarLL1(linea)
				data.Resultados = append(data.Resultados, Resultado{
					Cadena: linea,
					Pasos:  pasos,
					Valida: valida,
				})
				if valida {
					data.Validas++
				}
			}
		}
		tmpl.Execute(w, data)
	})

	fmt.Println("Metodo 2 - LL(1) No Recursivo → http://localhost:8082")
	http.ListenAndServe(":8082", nil)
}
