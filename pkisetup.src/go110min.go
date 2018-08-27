//+build !go1.10
// Add a canary file to suggest Go 1.10 is required 

/*
  @pkisetup go110min.go
  main.go       	version 1.0   created August 8, 2018

  Alain Pulluelo, VP Security & Mobile Innovation (alain.pulluelo@forgerock.com)

  ForgeRock Office of the CTO

  201 Mission St, Suite 2900
  San Francisco, CA 94105, USA
  +1(415)-559-1100

  Copyright (c) 2018, ForgeRock

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package main

// This file is here to give a better hint in the error message
// when this project is built with a too old version of Go.

var _ = ThisProjectRequiresGo1·10OrHigher
