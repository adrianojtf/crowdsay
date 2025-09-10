package poll

import (
	"fmt"
	"testing"
)

func TestVotePollDistribuicao(t *testing.T) {
	p := &Poll{
		Question: "Qual fruta?",
		Options:  []string{"Banana", "Maçã", "Uva"},
	}
	_ = CreatePoll(p)
	total := 100
	// 50 votos para Banana, 25 para Maçã, 25 para Uva
	for i := 0; i < 50; i++ {
		ip := fmt.Sprintf("10.0.0.%d", i)
		err := VotePoll(p.ID, "Banana", ip)
		if err != nil {
			t.Fatalf("erro ao votar em Banana: %v", err)
		}
	}
	for i := 50; i < 75; i++ {
		ip := fmt.Sprintf("10.0.0.%d", i)
		err := VotePoll(p.ID, "Maçã", ip)
		if err != nil {
			t.Fatalf("erro ao votar em Maçã: %v", err)
		}
	}
	for i := 75; i < 100; i++ {
		ip := fmt.Sprintf("10.0.0.%d", i)
		err := VotePoll(p.ID, "Uva", ip)
		if err != nil {
			t.Fatalf("erro ao votar em Uva: %v", err)
		}
	}
	if p.Results["Banana"] != 50 {
		t.Errorf("esperado 50 votos para Banana, obteve %d", p.Results["Banana"])
	}
	if p.Results["Maçã"] != 25 {
		t.Errorf("esperado 25 votos para Maçã, obteve %d", p.Results["Maçã"])
	}
	if p.Results["Uva"] != 25 {
		t.Errorf("esperado 25 votos para Uva, obteve %d", p.Results["Uva"])
	}
	if (p.Results["Banana"] + p.Results["Maçã"] + p.Results["Uva"]) != total {
		t.Errorf("total de votos deveria ser %d", total)
	}
}

func TestCreatePollAndGetPoll(t *testing.T) {
	p := &Poll{
		Question: "Qual sua linguagem favorita?",
		Options:  []string{"Go", "Python", "JavaScript"},
	}
	err := CreatePoll(p)
	if err != nil {
		t.Fatalf("erro ao criar enquete: %v", err)
	}
	got, err := GetPoll(p.ID)
	if err != nil {
		t.Fatalf("erro ao buscar enquete: %v", err)
	}
	if got.Question != p.Question {
		t.Errorf("esperado %q, obteve %q", p.Question, got.Question)
	}
}

func TestVotePoll(t *testing.T) {
	p := &Poll{
		Question: "Melhor cor?",
		Options:  []string{"Azul", "Verde"},
	}
	_ = CreatePoll(p)
	ip := "1.2.3.4"
	err := VotePoll(p.ID, "Azul", ip)
	if err != nil {
		t.Fatalf("erro ao votar: %v", err)
	}
	if p.Results["Azul"] != 1 {
		t.Errorf("esperado 1 voto para Azul, obteve %d", p.Results["Azul"])
	}
	// Não permite votar novamente do mesmo IP
	err = VotePoll(p.ID, "Azul", ip)
	if err == nil {
		t.Error("esperado erro de voto duplicado por IP")
	}
}

func TestVotePollOptionInvalida(t *testing.T) {
	p := &Poll{
		Question: "Melhor animal?",
		Options:  []string{"Cachorro", "Gato"},
	}
	_ = CreatePoll(p)
	err := VotePoll(p.ID, "Papagaio", "8.8.8.8")
	if err == nil {
		t.Error("esperado erro para opção inválida")
	}
}

func TestHasVotedIP(t *testing.T) {
	p := &Poll{
		Question: "Melhor estação?",
		Options:  []string{"Verão", "Inverno"},
	}
	_ = CreatePoll(p)
	ip := "5.5.5.5"
	if HasVotedIP(p, ip) {
		t.Error("não deveria ter voto registrado ainda")
	}
	_ = VotePoll(p.ID, "Verão", ip)
	if !HasVotedIP(p, ip) {
		t.Error("deveria detectar voto por IP")
	}
}
