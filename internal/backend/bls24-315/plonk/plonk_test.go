// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gnark DO NOT EDIT

package plonk_test

import (
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr"

	curve "github.com/consensys/gnark-crypto/ecc/bls24-315"

	"github.com/consensys/gnark/internal/backend/bls24-315/cs"

	bls24_315witness "github.com/consensys/gnark/internal/backend/bls24-315/witness"

	bls24_315plonk "github.com/consensys/gnark/internal/backend/bls24-315/plonk"

	"bytes"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr/kzg"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/plonk"
)

//--------------------//
//     benches		  //
//--------------------//

type refCircuit struct {
	nbConstraints int
	X             frontend.Variable
	Y             frontend.Variable `gnark:",public"`
}

func (circuit *refCircuit) Define(api frontend.API) error {
	for i := 0; i < circuit.nbConstraints; i++ {
		circuit.X = api.Mul(circuit.X, circuit.X)
	}
	api.AssertIsEqual(circuit.X, circuit.Y)
	return nil
}

func referenceCircuit() (frontend.CompiledConstraintSystem, frontend.Circuit, *kzg.SRS) {
	const nbConstraints = 40000
	circuit := refCircuit{
		nbConstraints: nbConstraints,
	}
	ccs, err := frontend.Compile(curve.ID, backend.UNKNOWN, &circuit, frontend.WithBuilder(plonk.NewBuilder))
	if err != nil {
		panic(err)
	}

	var good refCircuit
	good.X = (2)

	// compute expected Y
	var expectedY fr.Element
	expectedY.SetUint64(2)

	for i := 0; i < nbConstraints; i++ {
		expectedY.Mul(&expectedY, &expectedY)
	}

	good.Y = (expectedY)
	srs, err := kzg.NewSRS(ecc.NextPowerOfTwo(nbConstraints)+3, new(big.Int).SetUint64(42))
	if err != nil {
		panic(err)
	}

	return ccs, &good, srs
}

func BenchmarkSetup(b *testing.B) {
	ccs, _, srs := referenceCircuit()

	b.ResetTimer()

	b.Run("setup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = bls24_315plonk.Setup(ccs.(*cs.SparseR1CS), srs)
		}
	})
}

func BenchmarkProver(b *testing.B) {
	ccs, _solution, srs := referenceCircuit()
	fullWitness := bls24_315witness.Witness{}
	_, err := fullWitness.FromAssignment(_solution, false)
	if err != nil {
		b.Fatal(err)
	}

	pk, _, err := bls24_315plonk.Setup(ccs.(*cs.SparseR1CS), srs)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = bls24_315plonk.Prove(ccs.(*cs.SparseR1CS), pk, fullWitness, backend.ProverOption{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerifier(b *testing.B) {
	ccs, _solution, srs := referenceCircuit()
	fullWitness := bls24_315witness.Witness{}
	_, err := fullWitness.FromAssignment(_solution, false)
	if err != nil {
		b.Fatal(err)
	}
	publicWitness := bls24_315witness.Witness{}
	_, err = publicWitness.FromAssignment(_solution, true)
	if err != nil {
		b.Fatal(err)
	}

	pk, vk, err := bls24_315plonk.Setup(ccs.(*cs.SparseR1CS), srs)
	if err != nil {
		b.Fatal(err)
	}

	proof, err := bls24_315plonk.Prove(ccs.(*cs.SparseR1CS), pk, fullWitness, backend.ProverOption{})
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bls24_315plonk.Verify(proof, vk, publicWitness)
	}
}

func BenchmarkSerialization(b *testing.B) {
	ccs, _solution, srs := referenceCircuit()
	fullWitness := bls24_315witness.Witness{}
	_, err := fullWitness.FromAssignment(_solution, false)
	if err != nil {
		b.Fatal(err)
	}

	pk, _, err := bls24_315plonk.Setup(ccs.(*cs.SparseR1CS), srs)
	if err != nil {
		b.Fatal(err)
	}

	proof, err := bls24_315plonk.Prove(ccs.(*cs.SparseR1CS), pk, fullWitness, backend.ProverOption{})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	// ---------------------------------------------------------------------------------------------
	// bls24_315plonk.ProvingKey binary serialization
	b.Run("pk: binary serialization (bls24_315plonk.ProvingKey)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			_, _ = pk.WriteTo(&buf)
		}
	})
	b.Run("pk: binary deserialization (bls24_315plonk.ProvingKey)", func(b *testing.B) {
		var buf bytes.Buffer
		_, _ = pk.WriteTo(&buf)
		var pkReconstructed bls24_315plonk.ProvingKey
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(buf.Bytes())
			_, _ = pkReconstructed.ReadFrom(buf)
		}
	})
	{
		var buf bytes.Buffer
		_, _ = pk.WriteTo(&buf)
	}

	// ---------------------------------------------------------------------------------------------
	// bls24_315plonk.Proof binary serialization
	b.Run("proof: binary serialization (bls24_315plonk.Proof)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			_, _ = proof.WriteTo(&buf)
		}
	})
	b.Run("proof: binary deserialization (bls24_315plonk.Proof)", func(b *testing.B) {
		var buf bytes.Buffer
		_, _ = proof.WriteTo(&buf)
		var proofReconstructed bls24_315plonk.Proof
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(buf.Bytes())
			_, _ = proofReconstructed.ReadFrom(buf)
		}
	})
	{
		var buf bytes.Buffer
		_, _ = proof.WriteTo(&buf)
	}

}
