package srcbuilder

import (
	"fmt"
	"github.com/sirkon/ldetool/internal/ast"
	"github.com/sirkon/message"
)

func (sb *SrcBuilder) DispatchAnonymousOption(a *ast.AnonymousOption) error {
	sb.anonDepth++
	sb.appendGens(func() error {
		return sb.gen.OpenOptionalScope("", a.StartToken)
	})
	if err := sb.composeRule(a.Actions); err != nil {
		return err
	}
	sb.appendGens(func() error {
		return sb.gen.CloseOptionalScope()
	})
	message.Infof("End of anonymous option")

	sb.anonDepth--
	return nil
}

func (sb *SrcBuilder) DispatchAtEnd(a ast.AtEnd) error {
	if err := sb.gen.RegGravity(sb.prefixCur()); err != nil {
		return err
	}
	sb.appendGens(func() error {
		return sb.gen.AtEnd()
	})
	return nil
}

func (sb *SrcBuilder) DispatchErrorMismatch(a ast.ErrorOnMismatch) error {
	sb.appendGens(sb.gen.Stress)
	return nil
}

func (sb *SrcBuilder) DispatchMayBeStartChar(a *ast.MayBeStartChar) error {
	if err := sb.gen.RegGravity(sb.prefixCur()); err != nil {
		return err
	}
	sb.appendGens(func() error {
		return sb.gen.HeadChar(a.Value, true)
	})
	return nil
}

func (sb *SrcBuilder) DispatchMayBeStartString(a *ast.MayBeStartString) error {
	if err := sb.gen.RegGravity(sb.prefixCur()); err != nil {
		return err
	}
	sb.appendGens(func() error {
		return sb.gen.HeadString(a.Value, true)
	})
	return nil
}

func (sb *SrcBuilder) DispatchOptional(a *ast.Optional) error {
	if sb.anonDepth > 0 {
		return fmt.Errorf(
			"%d:%d: cannot create named optional area in anonymous one",
			a.NameToken.GetLine(),
			a.NameToken.GetColumn()+2,
		)
	}
	gotifiedName := sb.gotify.Public(a.Name)
	if gotifiedName != a.Name {
		sb.errToken = a.NameToken
		return fmt.Errorf("Wrong option identifier %s, must be %s", a.Name, gotifiedName)
	}
	sb.appendGens(func() error {
		return sb.gen.OpenOptionalScope(a.Name, a.NameToken)
	})
	sb.prefixTie(a.Name)
	defer sb.prefixUntie()
	if err := sb.composeRule(a.Actions); err != nil {
		return err
	}
	message.Infof("End of option \033[1m%s\033[0m", a.Name)
	sb.appendGens(sb.gen.CloseOptionalScope)
	return nil
}

func (sb *SrcBuilder) DispatchPassFirst(a ast.PassFixed) error {
	if err := sb.gen.RegGravity(sb.prefixCur()); err != nil {
		return err
	}
	sb.appendGens(func() error {
		return sb.gen.PassN(int(a))
	})
	return nil
}

func (sb *SrcBuilder) DispatchPassUntil(a *ast.PassUntil) error {
	if err := sb.gen.RegGravity(sb.prefixCur()); err != nil {
		return err
	}
	l := a.Limit
	var lower int
	var upper int
	lower = l.Lower
	upper = l.Upper

	if lower == upper && lower > 0 {
		// Fixed position check
		switch l.Type {
		case ast.String:
			sb.appendGens(func() error {
				return sb.gen.LookupFixedString(l.Value, lower, false)
			})
		case ast.Char:
			sb.appendGens(func() error {
				return sb.gen.LookupFixedChar(l.Value, lower, false)
			})
		default:
			return fmt.Errorf("fatal flow: passing action integrity error, got unexpected type %s", l.Type)
		}
	} else {
		// It is either short or limited/bounded lookup
		switch l.Type {
		case ast.String:
			sb.appendGens(func() error {
				return sb.gen.LookupString(l.Value, lower, upper, l.Close, false)
			})
		case ast.Char:
			sb.appendGens(func() error {
				return sb.gen.LookupChar(l.Value, lower, upper, l.Close, false)
			})
		default:
			return fmt.Errorf("fatal flow: passing action integrity error, got unexpected type %s", l.Type)
		}
	}
	return nil
}

func (sb *SrcBuilder) DispatchPassUntilOrIgnore(a *ast.PassUntilOrIgnore) error {
	if err := sb.gen.RegGravity(sb.prefixCur()); err != nil {
		return err
	}
	l := a.Limit
	var lower int
	var upper int
	lower = l.Lower
	upper = l.Upper

	if lower == upper && lower > 0 {
		// Fixed position check
		switch l.Type {
		case ast.String:
			sb.appendGens(func() error {
				return sb.gen.LookupFixedString(l.Value, lower, true)
			})
		case ast.Char:
			sb.appendGens(func() error {
				return sb.gen.LookupFixedChar(l.Value, lower, true)
			})
		}
	} else {
		// It is either short or limited/bounded lookup
		switch l.Type {
		case ast.String:
			sb.appendGens(func() error {
				return sb.gen.LookupString(l.Value, lower, upper, l.Close, true)
			})
		case ast.Char:
			sb.appendGens(func() error {
				return sb.gen.LookupChar(l.Value, lower, upper, l.Close, true)
			})
		}
	}
	return nil
}

func (sb *SrcBuilder) DispatchStartChar(a *ast.StartChar) error {
	if err := sb.gen.RegGravity(sb.prefixCur()); err != nil {
		return err
	}
	sb.appendGens(func() error {
		return sb.gen.HeadChar(a.Value, false)
	})
	return nil
}

func (sb *SrcBuilder) DispatchStartString(a *ast.StartString) error {
	if err := sb.gen.RegGravity(sb.prefixCur()); err != nil {
		return err
	}
	sb.appendGens(func() error {
		return sb.gen.HeadString(a.Value, false)
	})
	return nil
}

func (sb *SrcBuilder) DispatchTake(a *ast.Take) error {
	if sb.anonDepth > 0 {
		return fmt.Errorf(
			"%d:%d: cannot take while being in anonymous area",
			a.Field.NameToken.GetLine(),
			a.Field.NameToken.GetColumn()+2,
		)
	}
	if err := sb.checkField(a.Field); err != nil {
		return err
	}
	sb.prefixTie(a.Field.Name)
	defer sb.prefixUntie()
	if err := sb.gen.RegGravity(sb.prefixCur()); err != nil {
		return err
	}
	lower := a.Limit.Lower
	upper := a.Limit.Upper
	switch a.Limit.Type {
	case ast.String:
		sb.appendGens(func() error {
			if err := sb.gen.AddField(a.Field.Name, a.Field.Type, a.Field.NameToken); err != nil {
				return err
			}
			if err := sb.gen.TakeBeforeString(
				a.Field.Name, a.Field.Type, a.Limit.Value, lower, upper,
				a.Limit.Close, false); err != nil {
				return err
			}
			return nil
		})
	case ast.Char:
		sb.appendGens(func() error {
			if err := sb.gen.AddField(a.Field.Name, a.Field.Type, a.Field.NameToken); err != nil {
				return err
			}
			if err := sb.gen.TakeBeforeChar(
				a.Field.Name, a.Field.Type, a.Limit.Value, lower, upper,
				a.Limit.Close, false); err != nil {
				return err
			}
			return nil
		})
	}
	return nil
}

func (sb *SrcBuilder) DispatchTakeRest(a *ast.TakeRest) error {
	if sb.anonDepth > 0 {
		return fmt.Errorf(
			"%d:%d: cannot take the rest while being in anonymous area",
			a.Field.NameToken.GetLine(),
			a.Field.NameToken.GetColumn()+2,
		)
	}
	if err := sb.checkField(a.Field); err != nil {
		return err
	}
	sb.prefixTie(a.Field.Name)
	defer sb.prefixUntie()
	if err := sb.gen.RegGravity(sb.prefixCur()); err != nil {
		return err
	}
	sb.appendGens(func() error {
		if err := sb.gen.AddField(a.Field.Name, a.Field.Type, a.Field.NameToken); err != nil {
			return err
		}
		if err := sb.gen.TakeRest(a.Field.Name, a.Field.Type); err != nil {
			return err
		}
		return nil
	})
	return nil
}

func (sb *SrcBuilder) DispatchTakeUntilOrRest(a *ast.TakeUntilOrRest) error {
	if sb.anonDepth > 0 {
		return fmt.Errorf(
			"%d:%d: cannot take in anonymous area",
			a.Field.NameToken.GetLine(),
			a.Field.NameToken.GetColumn()+2,
		)
	}
	if err := sb.checkField(a.Field); err != nil {
		return err
	}
	sb.prefixTie(a.Field.Name)
	defer sb.prefixUntie()
	if err := sb.gen.RegGravity(sb.prefixCur()); err != nil {
		return err
	}
	switch a.Limit.Type {
	case ast.String:
		sb.appendGens(func() error {
			if err := sb.gen.AddField(a.Field.Name, a.Field.Type, a.Field.NameToken); err != nil {
				return err
			}
			if err := sb.gen.TakeBeforeString(
				a.Field.Name, a.Field.Type, a.Limit.Value, a.Limit.Lower, a.Limit.Upper,
				a.Limit.Close, true); err != nil {
				return err
			}
			return nil
		})
	case ast.Char:
		sb.appendGens(func() error {
			if err := sb.gen.AddField(a.Field.Name, a.Field.Type, a.Field.NameToken); err != nil {
				return err
			}
			if err := sb.gen.TakeBeforeChar(
				a.Field.Name, a.Field.Type, a.Limit.Value, a.Limit.Lower, a.Limit.Upper,
				a.Limit.Close, true); err != nil {
				return err
			}
			return nil
		})
	}
	return nil
}

func (sb *SrcBuilder) DispatchRule(a *ast.Rule) error {
	if err := sb.gen.UseRule(a.Name, a.NameToken); err != nil {
		return err
	}
	sb.generators = nil
	if err := sb.composeRule(a.Actions); err != nil {
		return err
	}
	return nil
}
