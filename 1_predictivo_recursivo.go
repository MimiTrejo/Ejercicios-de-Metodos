package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

// ============================================================
// METODO 1: DESCENDENTE PREDICTIVO RECURSIVO
//
// Gramatica:
//   S → a S b
//   S → a b
//
// Equivalente sin recursion izquierda (ya esta bien):
//   Al ver 'a', se consume, luego se decide:
//     si el siguiente es 'a' → aplicar S → a S b
//     si el siguiente es 'b' → aplicar S → a b
// ============================================================

// ANALIZADOR LEXICO
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

// PARSER RECURSIVO
type Parser struct {
	tokens []rune
	pos    int
	pasos  []string
}

func (p *Parser) actual() rune {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return '$'
}

func (p *Parser) consumir(esperado rune) error {
	if p.actual() == esperado {
		p.pasos = append(p.pasos, fmt.Sprintf("  consumir('%c')", esperado))
		p.pos++
		return nil
	}
	return fmt.Errorf("se esperaba '%c' pero llego '%c'", esperado, p.actual())
}

// S → a S b | a b
func (p *Parser) parseS() error {
	if p.actual() != 'a' {
		return fmt.Errorf("se esperaba 'a' pero llego '%c'", p.actual())
	}

	// lookahead: ver que hay despues de consumir 'a'
	siguiente := '$'
	if p.pos+1 < len(p.tokens) {
		siguiente = p.tokens[p.pos+1]
	}

	if siguiente == 'a' {
		// S → a S b
		p.pasos = append(p.pasos, "S → a S b")
		if err := p.consumir('a'); err != nil {
			return err
		}
		if err := p.parseS(); err != nil {
			return err
		}
		return p.consumir('b')
	} else if siguiente == 'b' {
		// S → a b
		p.pasos = append(p.pasos, "S → a b")
		if err := p.consumir('a'); err != nil {
			return err
		}
		return p.consumir('b')
	}

	return fmt.Errorf("produccion no encontrada para S con siguiente='%c'", siguiente)
}

// FUNCION PRINCIPAL
func analizar(cadena string) ([]string, bool) {
	cadena = strings.TrimSpace(cadena)
	if cadena == "" {
		return nil, false
	}

	tokens, err := lexer(cadena)
	if err != nil {
		return []string{"Error Lexico: " + err.Error()}, false
	}

	p := &Parser{tokens: tokens}

	if err := p.parseS(); err != nil {
		p.pasos = append(p.pasos, "Error Sintactico: "+err.Error())
		return p.pasos, false
	}

	if p.actual() != '$' {
		p.pasos = append(p.pasos, fmt.Sprintf("Error: sobran caracteres desde '%c'", p.actual()))
		return p.pasos, false
	}

	p.pasos = append(p.pasos, "VALIDA")
	return p.pasos, true
}

// PLANTILLA HTML
const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Metodo 1 - Predictivo Recursivo</title>
</head>
<body>

<h1>Metodo 1: Descendente Predictivo Recursivo</h1>

<pre>
Gramatica:
  S → a S b
  S → a b

Ejemplos validos: ab, aabb, aaabbb
</pre>

<form method="POST">
<label>Ingrese cadenas (una por linea):</label><br><br>
<textarea name="cadenas" rows="6" cols="30">{{.Input}}</textarea><br><br>
<button type="submit">Analizar</button>
</form>

{{if .Resultados}}
<h2>Resultados</h2>
<ul>
{{range .Resultados}}
<li>
  <b>{{.Cadena}}</b> → {{if .Valida}}VALIDA{{else}}INVALIDA{{end}}<br>
  <small>{{range .Pasos}}{{.}}<br>{{end}}</small>
</li>
<br>
{{end}}
</ul>
<p>Total validas: {{.Validas}} / {{len .Resultados}}</p>
{{end}}

</body>
</html>
`

type Resultado struct {
	Cadena string
	Pasos  []string
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
				pasos, valida := analizar(linea)
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

	fmt.Println("Metodo 1 - Predictivo Recursivo → http://localhost:8081")
	http.ListenAndServe(":8081", nil)
}
