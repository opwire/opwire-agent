package handlers

import (
	"bytes"
	"io"
	"os/exec"
)

type PipeChain struct {
}

func (p *PipeChain) Run(ib *bytes.Buffer, ob *bytes.Buffer, eb *bytes.Buffer, chain ...*exec.Cmd) (err error) {
	pipes := make([]*io.PipeWriter, len(chain)-1)
	i := 0
	chain[i].Stdin = ib
	for ; i < len(chain)-1; i++ {
		ip, op := io.Pipe()
		pipes[i] = op
		chain[i].Stdout = op
		chain[i].Stderr = eb
		chain[i+1].Stdin = ip
	}
	chain[i].Stdout = ob
	chain[i].Stderr = eb

	if err := p.next(chain, pipes); err != nil {
		// log or do something with this error
	}
	return err
}

func (p *PipeChain) next(chain []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
	if chain[0].Process == nil {
		if err = chain[0].Start(); err != nil {
			return err
		}
	}
	if len(chain) > 1 {
		if err = chain[1].Start(); err != nil {
			return err
		}
		defer func() {
			if err == nil {
				pipes[0].Close()
				err = p.next(chain[1:], pipes[1:])
			}
		}()
	}
	return chain[0].Wait()
}
