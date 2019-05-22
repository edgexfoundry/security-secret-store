//+build !go1.10
// Add a canary file to suggest Go 1.10 is required 

/*
   Copyright 2018 ForgeRock AS.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

  @author: Alain Pulluelo, ForgeRock (created: August 8, 2018)
  @author: Tingyu Zeng, DELL (updated: May 21, 2019)
  @version: 1.0.0
*/

package pkisetup

// This file is here to give a better hint in the error message
// when this project is built with a too old version of Go.

var _ = ThisProjectRequiresGo1·10OrHigher
