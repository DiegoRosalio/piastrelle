package main

import (
	"bufio"
	"container/list"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Direzione int

const (
	Nord Direzione = iota
	NordEst
	Est
	SudEst
	Sud
	SudOvest
	Ovest
	NordOvest
	NumeroDirezioni
)

func (d Direzione) opposta() Direzione {
	return (d + 4) % NumeroDirezioni
}

type Punto struct {
	x, y int
}

func (p Punto) puntoA(d Direzione) Punto {
	switch d {
	case NordOvest, Nord, NordEst:
		p.y++
	case SudOvest, Sud, SudEst:
		p.y--
	}
	switch d {
	case NordOvest, Ovest, SudOvest:
		p.x--
	case NordEst, Est, SudEst:
		p.x++
	}
	return p
}

type Piastrella struct {
	Colore    string
	Intensità int
	adiacente [NumeroDirezioni]*Piastrella
}

func (p *Piastrella) statoIntorno() map[string]int {
	intorno := make(map[string]int, 8)
	for dir := Nord; dir <= NordOvest; dir++ {
		if p.adiacente[dir] != nil {
			intorno[p.adiacente[dir].Colore]++
		}
	}
	return intorno
}

func (p *Piastrella) String() string {
	return fmt.Sprintf("%s %d", p.Colore, p.Intensità)
}

type RequisitoRegola struct {
	Colore string
	Minimo int
}

func (r RequisitoRegola) String() string {
	return fmt.Sprintf("%d %s", r.Minimo, r.Colore)
}

type Regola struct {
	Requisiti []RequisitoRegola
	Risultato string
	hit       int
}

func (r *Regola) applicabile(intorno map[string]int) bool {
	for _, requisito := range r.Requisiti {
		if intorno[requisito.Colore] < requisito.Minimo {
			return false
		}
	}
	return true
}

func (r *Regola) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "%s:", r.Risultato)
	for _, req := range r.Requisiti {
		fmt.Fprintf(sb, " %s", req)
	}
	return sb.String()
}

type piano struct {
	piastrelle map[Punto]*Piastrella
	regole     *[]*Regola
}

func (p piano) aggiungiPiastrella(pos Punto, pst Piastrella) {
	p.piastrelle[pos] = &pst
	for dir := Nord; dir <= NordOvest; dir++ {
		if pstVicina, accesa := p.piastrelle[pos.puntoA(dir)]; accesa {
			pst.adiacente[dir] = pstVicina
			pstVicina.adiacente[dir.opposta()] = &pst
		}
	}
}

func (p piano) rimuoviPiastrella(pos Punto) {
	pst, accesa := p.piastrelle[pos]
	if !accesa {
		return
	}
	for dir := Nord; dir <= NordOvest; dir++ {
		if pst.adiacente[dir] == nil {
			continue
		}
		pst.adiacente[dir].adiacente[dir.opposta()] = nil
		pst.adiacente[dir] = nil
	}
	delete(p.piastrelle, pos)
}

func (p piano) aggiungiRegola(r Regola) {
	r.hit = 0
	(*p.regole) = append((*p.regole), &r)
}

func (p piano) ordinaRegole() {
	slices.SortStableFunc(*p.regole, func(a, b *Regola) int {
		return a.hit - b.hit
	})
}

func (p piano) primaRegolaApplicabile(intorno map[string]int) *Regola {
	i := slices.IndexFunc(*p.regole, func(r *Regola) bool { return r.applicabile(intorno) })
	if i == -1 {
		return nil
	}
	return (*p.regole)[i]
}

func (p piano) statoIntorno(pos Punto) map[string]int {
	pst, accesa := p.piastrelle[pos]
	if accesa {
		return pst.statoIntorno()
	}
	intorno := make(map[string]int, 8)
	for dir := Nord; dir <= NordOvest; dir++ {
		pstAdiacente, pstAdiacenteAccesa := p.piastrelle[pos.puntoA(dir)]
		if pstAdiacenteAccesa {
			intorno[pstAdiacente.Colore]++
		}
	}
	return intorno
}

func (p piano) propaga(pos Punto) {
	pst, accesa := p.piastrelle[pos]
	intorno := p.statoIntorno(pos)
	regola := p.primaRegolaApplicabile(intorno)
	if regola == nil {
		return
	}
	regola.hit++
	if !accesa {
		p.aggiungiPiastrella(pos, Piastrella{Colore: regola.Risultato, Intensità: 1})
	} else {
		pst.Colore = regola.Risultato
	}
}

func (p piano) propagaBlocco(pos Punto) {
	pst, accesa := p.piastrelle[pos]
	if !accesa {
		return
	}
	var piastrelleBlocco []*Piastrella
	p.visitaInAmpiezza(pst, func(p *Piastrella, _ int) bool {
		piastrelleBlocco = append(piastrelleBlocco, p)
		return true
	})
	risultatoPropagazione := make([]string, len(piastrelleBlocco))
	for i, piastrella := range piastrelleBlocco {
		regola := p.primaRegolaApplicabile(piastrella.statoIntorno())
		if regola == nil {
			continue
		}
		risultatoPropagazione[i] = regola.Risultato
		regola.hit++
	}
	for i, piastrella := range piastrelleBlocco {
		nuovoColore := risultatoPropagazione[i]
		if nuovoColore == "" {
			continue
		}
		piastrella.Colore = nuovoColore
	}
}

func (p piano) calcolaIntensitàBlocco(pnt Punto, omogeneo bool) int {
	pst, accesa := p.piastrelle[pnt]
	if !accesa {
		return 0
	}
	var intensitàTotale int
	colore := pst.Colore
	p.visitaInAmpiezza(pst, func(p *Piastrella, _ int) bool {
		if omogeneo && colore != p.Colore {
			return false
		}
		intensitàTotale += p.Intensità
		return true
	})
	return intensitàTotale
}

func (p piano) pista(pos Punto, direzioni []Direzione) []*Piastrella {
	pst, accesa := p.piastrelle[pos]
	if !accesa {
		return nil
	}
	piastrelle := make([]*Piastrella, 0, len(direzioni)+1)
	piastrelle = append(piastrelle, pst)
	for ; len(direzioni) != 0; direzioni = direzioni[1:] {
		pos = pos.puntoA(direzioni[0])
		pst = pst.adiacente[direzioni[0]]
		if pst == nil {
			break
		}
		piastrelle = append(piastrelle, pst)
	}
	if pst == nil {
		return nil
	}
	return piastrelle
}

func (p piano) lunghezzaPistaBreve(partenza, arrivo Punto) int {
	lunghezza := -1
	pstPartenza, accesa := p.piastrelle[partenza]
	if !accesa {
		return lunghezza
	}
	pstArrivo, accesa := p.piastrelle[arrivo]
	if !accesa {
		return lunghezza
	}
	p.visitaInAmpiezza(pstPartenza, func(pst *Piastrella, profondità int) bool {
		if lunghezza != -1 {
			return false
		}
		if pst == pstArrivo {
			lunghezza = profondità + 1
			return false
		}
		return true
	})
	return lunghezza
}

type Visitatore func(pst *Piastrella, profondità int) bool

func (p piano) visitaInAmpiezza(partenza *Piastrella, visita Visitatore) {
	if partenza == nil {
		return
	}
	statoEsplorazione := map[*Piastrella]bool{partenza: true}
	type elementoFrangia struct {
		*Piastrella
		profondità int
	}
	coda := list.New()
	coda.PushBack(elementoFrangia{Piastrella: partenza, profondità: 0})
	for coda.Len() != 0 {
		pst := coda.Remove(coda.Front()).(elementoFrangia)
		continuaVisita := visita(pst.Piastrella, pst.profondità)
		if !continuaVisita {
			continue
		}
		for dir := Nord; dir <= NordOvest; dir++ {
			if pst.adiacente[dir] == nil || statoEsplorazione[pst.adiacente[dir]] {
				continue
			}
			coda.PushBack(elementoFrangia{
				Piastrella: pst.adiacente[dir],
				profondità: pst.profondità + 1,
			})
			statoEsplorazione[pst.adiacente[dir]] = true
		}
	}
}

func colora(p piano, x int, y int, alpha string, i int) {
	p.aggiungiPiastrella(
		Punto{x: x, y: y},
		Piastrella{
			Colore:    alpha,
			Intensità: i,
		},
	)
}

func spegni(p piano, x int, y int) {
	p.rimuoviPiastrella(Punto{x: x, y: y})
}

func regola(p piano, r string) {
	parti := strings.Split(r, " ")
	reg := Regola{
		Risultato: parti[0],
		Requisiti: make([]RequisitoRegola, 0, (len(parti)-1)/2),
	}
	for i := 1; i < len(parti); i += 2 {
		m, _ := strconv.Atoi(parti[i])
		reg.Requisiti = append(reg.Requisiti, RequisitoRegola{
			Colore: parti[i+1],
			Minimo: int(m),
		})
	}
	p.aggiungiRegola(reg)
}

func stato(p piano, x int, y int) (string, int) {
	pnt := Punto{x: x, y: y}
	pst, accesa := p.piastrelle[pnt]
	if !accesa {
		return "", 0
	}
	fmt.Println(pst)
	return pst.Colore, pst.Intensità
}

func stampa(p piano) {
	fmt.Println("(")
	for _, r := range *p.regole {
		fmt.Println(r)
	}
	fmt.Println(")")
}

func esegui(p piano, s string) {
	args := strings.SplitN(s, " ", 2)
	comando, args := args[0], args[1:]

	switch comando {
	case "d":
		for pos, value := range p.piastrelle {
			fmt.Printf("%v %s %v\n", pos, value, value.adiacente)
		}
	case "dd":
		stampaMappa(p.piastrelle)
	case "r":
		regola(p, args[0])
		args = args[1:]
	case "o":
		p.ordinaRegole()
	case "s":
		stampa(p)
	case "q":
		os.Exit(0)
	}
	if len(args) == 0 {
		return
	}
	args = strings.Fields(args[0])
	x1, _ := strconv.Atoi(args[0])
	y1, _ := strconv.Atoi(args[1])
	args = args[2:]
	pos1 := Punto{x: x1, y: y1}

	switch comando {
	case "C":
		i, _ := strconv.Atoi(args[1])
		colora(p, x1, y1, args[0], i)
	case "S":
		spegni(p, x1, y1)
	case "?":
		stato(p, x1, y1)
	case "b", "B":
		i := p.calcolaIntensitàBlocco(pos1, comando == "B")
		fmt.Println(i)
	case "p":
		p.propaga(pos1)
	case "P":
		p.propagaBlocco(pos1)
	case "t":
		direzioni := parseDirezioni(args[0])
		pista := p.pista(pos1, direzioni)
		if len(pista) == 0 {
			break
		}
		fmt.Println("[")
		for i, pst := range pista {
			fmt.Printf("%d %d %s %d\n", pos1.x, pos1.y, pst.Colore, pst.Intensità)
			if i != len(pista)-1 {
				pos1 = pos1.puntoA(direzioni[i])
			}
		}
		fmt.Println("]")
	case "L":
		x2, _ := strconv.Atoi(args[0])
		y2, _ := strconv.Atoi(args[1])
		pos2 := Punto{x: x2, y: y2}
		if lg := p.lunghezzaPistaBreve(pos1, pos2); lg != -1 {
			fmt.Println(lg)
		}
	}
}

func parseDirezioni(str string) []Direzione {
	parti := strings.Split(str, ",")
	direzioni := make([]Direzione, len(parti))
	for i, parte := range parti {
		switch parte {
		case "NN":
			direzioni[i] = Nord
		case "NE":
			direzioni[i] = NordEst
		case "EE":
			direzioni[i] = Est
		case "SE":
			direzioni[i] = SudEst
		case "SS":
			direzioni[i] = Sud
		case "SO":
			direzioni[i] = SudOvest
		case "OO":
			direzioni[i] = Ovest
		case "NO":
			direzioni[i] = NordOvest
		}
	}
	return direzioni
}

func main() {
	p := piano{
		piastrelle: make(map[Punto]*Piastrella),
		regole:     &[]*Regola{},
	}
	for scanner := bufio.NewScanner(os.Stdin); scanner.Scan(); {
		esegui(p, scanner.Text())
	}
}

func stampaMappa(mappa map[Punto]*Piastrella) {
	maxX, maxY := 0, 0
	for p := range mappa {
		if p.x > maxX {
			maxX = p.x
		}
		if p.y > maxY {
			maxY = p.y
		}
	}

	griglia := make([][]string, maxY+1)
	for i := range griglia {
		griglia[i] = make([]string, maxX+1)
	}

	for p, piastrella := range mappa {
		if piastrella != nil {
			griglia[p.y][p.x] = piastrella.String()
		}
	}

	strs := make([]string, 0)
	for i := 0; i < maxX+1; i++ {
		strs = append(strs, "───")
	}
	topBorder := "┌" + strings.Join(strs, "┬") + "┐"
	midBorder := "├" + strings.Join(strs, "┼") + "┤"
	botBorder := "└" + strings.Join(strs, "┴") + "┘"

	fmt.Println(topBorder)
	for i := len(griglia) - 1; i >= 0; i-- {
		for _, cella := range griglia[i] {
			if len(cella) > 3 {
				cella = cella[:3]
			}
			paddingTot := 3 - len(cella)
			padLeft, padRight := paddingTot/2, (paddingTot+1)/2
			fmt.Printf("│%s%s%s", strings.Repeat(" ", padLeft), cella, strings.Repeat(" ", padRight))
		}
		fmt.Println("│")
		if i != 0 {
			fmt.Println(midBorder)
		}
	}
	fmt.Println(botBorder)
}
