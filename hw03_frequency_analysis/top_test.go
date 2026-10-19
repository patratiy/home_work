package hw03frequencyanalysis

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Change to true if needed.
var taskWithAsteriskIsCompleted = false

var text = `Как видите, он  спускается  по  лестнице  вслед  за  своим
	другом   Кристофером   Робином,   головой   вниз,  пересчитывая
	ступеньки собственным затылком:  бум-бум-бум.  Другого  способа
	сходить  с  лестницы  он  пока  не  знает.  Иногда ему, правда,
		кажется, что можно бы найти какой-то другой способ, если бы  он
	только   мог   на  минутку  перестать  бумкать  и  как  следует
	сосредоточиться. Но увы - сосредоточиться-то ему и некогда.
		Как бы то ни было, вот он уже спустился  и  готов  с  вами
	познакомиться.
	- Винни-Пух. Очень приятно!
		Вас,  вероятно,  удивляет, почему его так странно зовут, а
	если вы знаете английский, то вы удивитесь еще больше.
		Это необыкновенное имя подарил ему Кристофер  Робин.  Надо
	вам  сказать,  что  когда-то Кристофер Робин был знаком с одним
	лебедем на пруду, которого он звал Пухом. Для лебедя  это  было
	очень   подходящее  имя,  потому  что  если  ты  зовешь  лебедя
	громко: "Пу-ух! Пу-ух!"- а он  не  откликается,  то  ты  всегда
	можешь  сделать вид, что ты просто понарошку стрелял; а если ты
	звал его тихо, то все подумают, что ты  просто  подул  себе  на
	нос.  Лебедь  потом  куда-то делся, а имя осталось, и Кристофер
	Робин решил отдать его своему медвежонку, чтобы оно не  пропало
	зря.
		А  Винни - так звали самую лучшую, самую добрую медведицу
	в  зоологическом  саду,  которую  очень-очень  любил  Кристофер
	Робин.  А  она  очень-очень  любила  его. Ее ли назвали Винни в
	честь Пуха, или Пуха назвали в ее честь - теперь уже никто  не
	знает,  даже папа Кристофера Робина. Когда-то он знал, а теперь
	забыл.
		Словом, теперь мишку зовут Винни-Пух, и вы знаете почему.
		Иногда Винни-Пух любит вечерком во что-нибудь поиграть,  а
	иногда,  особенно  когда  папа  дома,  он больше любит тихонько
	посидеть у огня и послушать какую-нибудь интересную сказку.
		В этот вечер...`

func TestTop10(t *testing.T) {
	t.Run("no words in empty string", func(t *testing.T) {
		require.Len(t, Top10(""), 0)
	})

	t.Run("positive test", func(t *testing.T) {
		if taskWithAsteriskIsCompleted {
			expected := []string{
				"а",         // 8
				"он",        // 8
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"в",         // 4
				"его",       // 4
				"если",      // 4
				"кристофер", // 4
				"не",        // 4
			}
			require.Equal(t, expected, Top10(text))
		} else {
			expected := []string{
				"он",        // 8
				"а",         // 6
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"-",         // 4
				"Кристофер", // 4
				"если",      // 4
				"не",        // 4
				"то",        // 4
			}
			require.Equal(t, expected, Top10(text))
		}
	})
}

func TestTop10Additional(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "example from the task description",
			text:     "cat and dog, one dog,two cats and one man",
			expected: []string{"and", "one", "cat", "cats", "dog,", "dog,two", "man"},
		},
		{
			name:     "empty string",
			text:     "",
			expected: []string{},
		},
		{
			name:     "whitespace only",
			text:     " \t\n ",
			expected: []string{},
		},
		{
			name:     "single word",
			text:     "hello",
			expected: []string{"hello"},
		},
		{
			name:     "different letter case means different words",
			text:     "Кот кот Кот кот кот",
			expected: []string{"кот", "Кот"},
		},
		{
			name:     "punctuation is part of the word",
			text:     "кот, кот, кот! кот! кот кот",
			expected: []string{"кот", "кот!", "кот,"},
		},
		{
			name:     "dash is a word",
			text:     "a - b - c",
			expected: []string{"-", "a", "b", "c"},
		},
		{
			name:     "exactly ten different words",
			text:     "один два три четыре пять шесть семь восемь девять десять",
			expected: []string{"восемь", "два", "девять", "десять", "один", "пять", "семь", "три", "четыре", "шесть"},
		},
		{
			name:     "more than ten words with the same frequency",
			text:     strings.TrimSpace(strings.Repeat("a b c d e f g h i j k l m n o ", 133)),
			expected: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, Top10(tt.text))
		})
	}
}

func TestTop10Normalized(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "letter case is ignored",
			text:     "Нога нога НОГА нога",
			expected: []string{"нога"},
		},
		{
			name:     "punctuation at the word edges is removed",
			text:     "нога! нога, нога 'нога' нога.",
			expected: []string{"нога"},
		},
		{
			name:     "dash inside the word is kept",
			text:     "какой-то какойто какой-то",
			expected: []string{"какой-то", "какойто"},
		},
		{
			name:     "word made of dashes stays a word, a single dash does not",
			text:     "------- -",
			expected: []string{"-------"},
		},
		{
			name:     "punctuation inside the word makes words different",
			text:     "dog,cat dog...cat dogcat",
			expected: []string{"dog,cat", "dog...cat", "dogcat"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, Top10Normalized(tt.text))
		})
	}
}

func TestTop10NormalizedBigText(t *testing.T) {
	expected := []string{
		"а",
		"он",
		"и",
		"ты",
		"что",
		"в",
		"его",
		"если",
		"кристофер",
		"не",
	}
	require.Equal(t, expected, Top10Normalized(text))
}
