package config

import "errors"

type InvocationSemantic string

/*
  ┌──────────────────────┬────────────┬──────────────────┬────────────────────────────┐
  │  Invocation Semantic │ Retransmit │Filter Duplicates │Re-execute OR Re-transmit   │
  ├──────────────────────┼────────────┼──────────────────┼────────────────────────────┤
  │  Maybe               │     No     │       N.A.       │    N.A.                    │
  ├──────────────────────┼────────────┼──────────────────┼────────────────────────────┤
  │  At Least Once       │     Yes    │       No         │    Re-execute              │
  ├──────────────────────┼────────────┼──────────────────┼────────────────────────────┤
  │  At Most Once        │     Yes    │       Yes        │    Re-transmit             │
  └──────────────────────┴────────────┴──────────────────┴────────────────────────────┘
*/

const (
	Maybe       InvocationSemantic = "maybe"
	AtLeastOnce InvocationSemantic = "at-least-once"
	AtMostOnce  InvocationSemantic = "at-most-once"
)

func (s InvocationSemantic) Validate() error {
	switch s {
	case Maybe, AtLeastOnce, AtMostOnce:
		return nil
	default:
		return errors.New("invalid invocation semantic")
	}
}
