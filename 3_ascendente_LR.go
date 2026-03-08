package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

// ============================================================
// METODO 3: ASCENDENTE LR(0)/SLR(1)
//
// Gramatica aumentada:
//   G0: S' → S       (aumentada)
//   G1: S  → a S b
//   G2: S  → a b
//
// Estados LR:
//   I0: S' → .S,  S → .aSb,  S → .ab
//   I1: S' → S.                          (ACEPTAR con $)
//   I2: S  → a.Sb,  S → a.b,  S → .aSb, S → .ab
//   I3: S  → ab.                         (reduce G2)
//   I4: S  → aS.b
//   I5: S  → aSb.                        (reduce G1)
//
// Tabla ACTION/GOTO:
//   Estado | a    | b    | $    || S
//   -------+------+------+------++---
//     0    | s2   |      |      || 1
//     1    |      |      | acc  ||
//     2    | s2   | s3   |      || 4
//     3    |      | r2   | r2   ||
//     4    |      | s5   |      ||
//     5    |      | r1   | r1   ||
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

type Prod struct {
	Head string
	Len  int
	Desc string
}

var productions = []Prod{
	{},
	{"S", 3, "S → a S b"},
	{"S", 2, "S → a b"},
}

// ACTION[estado][terminal] → "sN" | "rN" | "acc" | ""
var actionTable = map[int]map[rune]string{
	0: {'a': "s2"},
	1: {'$': "acc"},
	2: {'a': "s2", 'b': "s3"},
	3: {'b': "r2", '$': "r2"},
	4: {'b': "s5"},
	5: {'b': "r1", '$': "r1"},
}

// GOTO[estado][no-terminal] → estado
var gotoTable = map[int]map[string]int{
	0: {"S": 1},
	2: {"S": 4},
}

type Row struct {
	Pila     string
	Simbolos string
	Entrada  string
	Accion   string
}

func analizarLR(cadena string) ([]Row, bool) {
	cadena = strings.TrimSpace(cadena)
	tokens, err := lexer(cadena)
	if err != nil {
		return []Row{{Accion: "Error Lexico: " + err.Error()}}, false
	}

	states := []int{0}
	symbols := []string{}
	ip := 0
	var rows []Row
	ok := true

	stackStr := func() string {
		parts := make([]string, len(states))
		for i, v := range states {
			parts[i] = fmt.Sprintf("%d", v)
		}
		return strings.Join(parts, " ")
	}
	symStr := func() string {
		if len(symbols) == 0 {
			return "$"
		}
		return "$ " + strings.Join(symbols, " ")
	}
	inputStr := func() string {
		var parts []string
		for _, t := range tokens[ip:] {
			parts = append(parts, string(t))
		}
		return strings.Join(parts, " ")
	}

	for {
		top := states[len(states)-1]
		cur := tokens[ip]
		act := actionTable[top][cur]

		rows = append(rows, Row{
			Pila:     stackStr(),
			Simbolos: symStr(),
			Entrada:  inputStr(),
			Accion:   act,
		})

		if act == "" {
			rows[len(rows)-1].Accion = fmt.Sprintf("Error: no hay accion en [%d]['%c']", top, cur)
			ok = false
			break
		}
		if act == "acc" {
			rows[len(rows)-1].Accion = "ACEPTAR"
			break
		}

		if act[0] == 's' {
			var nextState int
			fmt.Sscanf(act[1:], "%d", &nextState)
			states = append(states, nextState)
			symbols = append(symbols, string(cur))
			ip++
		} else if act[0] == 'r' {
			var prodNum int
			fmt.Sscanf(act[1:], "%d", &prodNum)
			p := productions[prodNum]
			states = states[:len(states)-p.Len]
			symbols = symbols[:len(symbols)-p.Len]
			symbols = append(symbols, p.Head)
			prevTop := states[len(states)-1]
			nextState := gotoTable[prevTop][p.Head]
			states = append(states, nextState)
			rows[len(rows)-1].Accion = "reduce: " + p.Desc
		}
	}
	return rows, ok
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Metodo 3 - LR Ascendente</title>
</head>
<body>

<h1>Metodo 3: Ascendente LR(0)/SLR(1)</h1>

<pre>
Gramatica aumentada:
  G0: S' → S
  G1: S  → a S b
  G2: S  → a b

Tabla ACTION/GOTO:
  Estado | a    | b    | $    || S
  -------+------+------+------++---
    0    | s2   |      |      || 1
    1    |      |      | acc  ||
    2    | s2   | s3   |      || 4
    3    |      | r2   | r2   ||
    4    |      | s5   |      ||
    5    |      | r1   | r1   ||

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
    <th>Pila de estados</th>
    <th>Simbolos</th>
    <th>Entrada</th>
    <th>Accion</th>
  </tr>
  {{range .Pasos}}
  <tr>
    <td>{{.Pila}}</td>
    <td>{{.Simbolos}}</td>
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
	Pasos  []Row
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
				pasos, valida := analizarLR(linea)
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

	fmt.Println("Metodo 3 - LR Ascendente → http://localhost:8083")
	http.ListenAndServe(":8083", nil)
}
