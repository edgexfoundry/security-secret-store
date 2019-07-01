/*
   Copyright 2019 DELL Technologies.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
  
  @author: Tingyu Zeng, DELL (created: May 21, 2019)
  @version: 1.0.0
*/

package vaultworker

import (
	"testing"
	"fmt"	
)

const TokenfilepathWin = "..\\..\\..\\test\\test-resp-init.json"
const TokenfilepathUnix = "../../../test/test-resp-init.json"
func TestGetSecret(t *testing.T){
	 p := TokenfilepathWin
	token, err := GetSecret(p)

	if err != nil {
		t.Errorf("Failed to get secret from file.")
	}

	if len(token.Token) < 1 {
		t.Errorf("Failed to get secret from file.")
	}

}

func TestGetSecretNoExistFile(t *testing.T) {
	token, err := GetSecret("\\no\\exist\\file")
	if err != nil {
		fmt.Printf(err.Error())
	}

	if len(token.Token) > 1 {
		t.Errorf("expected a nil token, instead having %s", token.Token)
	}
}
