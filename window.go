/*
Copyright © 2025 Matt Krueger <mkrueger@rstms.net>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice,
    this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice,
    this list of conditions and the following disclaimer in the documentation
    and/or other materials provided with the distribution.

 3. Neither the name of the copyright holder nor the names of its contributors
    may be used to endorse or promote products derived from this software
    without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package common

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

//go:embed conwinctl.exe
var conwinctlData []byte

func HideConsoleWindow() error {
	err := SetConsoleWindowVisibility("hide")
	if err != nil {
		return Fatal(err)
	}
	return nil
}

func SetConsoleWindowVisibility(visibility string) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	tempDir, err := os.MkdirTemp("", "conwinctl-*")
	if err != nil {
		return Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	conwinctlCommand := filepath.Join(tempDir, "conwinctl.exe")
	err = os.WriteFile(conwinctlCommand, conwinctlData, 0700)
	if err != nil {
		return Fatal(err)
	}
	err = exec.Command(conwinctlCommand, visibility).Run()
	if err != nil {
		return Fatal(err)
	}
	return nil
}
