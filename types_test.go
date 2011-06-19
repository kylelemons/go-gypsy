package goyaml

import (
	"testing"
)

var stringTests = []struct{
	Tree Node
	Expect string
}{
	{
		Tree: Scalar("test"),
		Expect: `test`,
	},
	{
		Tree: List{
			Scalar("One"),
			Scalar("Two"),
			Scalar("Three"),
		},
		Expect: `- One
- Two
- Three
`,
	},
	{
		Tree: Map{
			"phonetic": Scalar("true"),
			"organization": Scalar("Navy"),
			"alphabet": List{
				Scalar("Alpha"),
				Scalar("Bravo"),
				Scalar("Charlie"),
			},
		},
		Expect: `organization: Navy
phonetic:     true
alphabet:
  - Alpha
  - Bravo
  - Charlie
`,
	},
	{
		Tree: Map{
			"answer": Scalar("42"),
			"question": List{
				Scalar("What do you get when you multiply six by nine?"),
				Scalar("How many roads must a man walk down?"),
			},
		},
		Expect: `answer: 42
question:
  - What do you get when you multiply six by nine?
  - How many roads must a man walk down?
`,
	},
	{
		Tree: List{
			Map{
				"name": Scalar("John Smith"),
				"age": Scalar("42"),
			},
			Map{
				"name": Scalar("Jane Smith"),
				"age": Scalar("45"),
			},
		},
		Expect: `- age:  42
  name: John Smith
- age:  45
  name: Jane Smith
`,
	},
}

func TestString(t *testing.T) {
	for idx, test := range stringTests {
		if got, want := test.Tree.String(), test.Expect; got != want {
			t.Errorf("%d. got:\n%s\n%d. want:\n%s\n", idx, got, idx, want)
		}
	}
}
