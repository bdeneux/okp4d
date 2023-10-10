//nolint:gocognit,lll
package predicate

import (
	"fmt"
	"testing"

	"github.com/ichiban/prolog/engine"

	. "github.com/smartystreets/goconvey/convey"

	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/okp4/okp4d/x/logic/testutil"
	"github.com/okp4/okp4d/x/logic/types"
)

func TestCryptoHash(t *testing.T) {
	Convey("Given a test cases", t, func() {
		cases := []struct {
			program     string
			query       string
			wantResult  []types.TermResults
			wantError   error
			wantSuccess bool
		}{
			{
				query: `sha_hash('foo', Hash).`,
				wantResult: []types.TermResults{{
					"Hash": "[44,38,180,107,104,255,198,143,249,155,69,60,29,48,65,52,19,66,45,112,100,131,191,160,249,138,94,136,98,102,231,174]",
				}},
				wantSuccess: true,
			},
			{
				query:       `sha_hash(Foo, Hash).`,
				wantResult:  []types.TermResults{},
				wantError:   fmt.Errorf("sha_hash/2: invalid data type: engine.Variable, should be Atom"),
				wantSuccess: false,
			},
			{
				query: `sha_hash('bar',
[44,38,180,107,104,255,198,143,249,155,69,60,29,48,65,52,19,66,45,112,100,131,191,160,249,138,94,136,98,102,231,174]).`,
				wantSuccess: false,
			},
			{
				query: `sha_hash('bar',
[252,222,43,46,219,165,107,244,8,96,31,183,33,254,155,92,51,141,16,238,66,158,160,79,174,85,17,182,143,191,143,185]).`,
				wantResult:  []types.TermResults{{}},
				wantSuccess: true,
			},
			{
				query: `sha_hash('bar',
[345,222,43,46,219,165,107,244,8,96,31,183,33,254,155,92,51,141,16,238,66,158,160,79,174,85,17,182,143,191,143,185]).`,
				wantSuccess: false,
			},
			{
				program: `test :- sha_hash('bar', H),
H == [252,222,43,46,219,165,107,244,8,96,31,183,33,254,155,92,51,141,16,238,66,158,160,79,174,85,17,182,143,191,143,185].`,
				query:       `test.`,
				wantResult:  []types.TermResults{{}},
				wantSuccess: true,
			},
			{
				program: `test :- sha_hash('bar', H),
H == [2252,222,43,46,219,165,107,244,8,96,31,183,33,254,155,92,51,141,16,238,66,158,160,79,174,85,17,182,143,191,143,185].`,
				query:       `test.`,
				wantSuccess: false,
			},
			{
				query: `hex_bytes(Hex,
[44,38,180,107,104,255,198,143,249,155,69,60,29,48,65,52,19,66,45,112,100,131,191,160,249,138,94,136,98,102,231,174]).`,
				wantResult:  []types.TermResults{{"Hex": "'2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae'"}},
				wantSuccess: true,
			},
			{
				query: `hex_bytes('2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae', Bytes).`,
				wantResult: []types.TermResults{{
					"Bytes": "[44,38,180,107,104,255,198,143,249,155,69,60,29,48,65,52,19,66,45,112,100,131,191,160,249,138,94,136,98,102,231,174]",
				}},
				wantSuccess: true,
			},
			{
				query: `hex_bytes('2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae',
[44,38,180,107,104,255,198,143,249,155,69,60,29,48,65,52,19,66,45,112,100,131,191,160,249,138,94,136,98,102,231,174]).`,
				wantResult:  []types.TermResults{{}},
				wantSuccess: true,
			},
			{
				query: `hex_bytes('3c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae',
[44,38,180,107,104,255,198,143,249,155,69,60,29,48,65,52,19,66,45,112,100,131,191,160,249,138,94,136,98,102,231,174]).`,
				wantSuccess: false,
			},
			{
				query: `hex_bytes('fail',
[44,38,180,107,104,255,198,143,249,155,69,60,29,48,65,52,19,66,45,112,100,131,191,160,249,138,94,136,98,102,231,174]).`,
				wantError:   fmt.Errorf("hex_bytes/2: failed decode hexadecimal encoding/hex: invalid byte: U+0069 'i'"),
				wantSuccess: false,
			},
			{
				query: `hex_bytes('2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae',
[345,38,180,107,104,255,198,143,249,155,69,60,29,48,65,52,19,66,45,112,100,131,191,160,249,138,94,136,98,102,231,174]).`,
				wantSuccess: false,
			},
			{
				query: `hex_bytes('2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae',
[345,38,'hey',107,104,255,198,143,249,155,69,60,29,48,65,52,19,66,45,112,100,131,191,160,249,138,94,136,98,102,231,174]).`,
				wantSuccess: false,
				wantError:   fmt.Errorf("hex_bytes/2: failed convert list into bytes: invalid term type in list engine.Atom, only integer allowed"),
			},
		}
		for nc, tc := range cases {
			Convey(fmt.Sprintf("Given the query #%d: %s", nc, tc.query), func() {
				Convey("and a context", func() {
					db := tmdb.NewMemDB()
					stateStore := store.NewCommitMultiStore(db)
					ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

					Convey("and a vm", func() {
						interpreter := testutil.NewLightInterpreterMust(ctx)
						interpreter.Register2(engine.NewAtom("sha_hash"), SHAHash)
						interpreter.Register2(engine.NewAtom("hex_bytes"), HexBytes)

						err := interpreter.Compile(ctx, tc.program)
						So(err, ShouldBeNil)

						Convey("When the predicate is called", func() {
							sols, err := interpreter.QueryContext(ctx, tc.query)

							Convey("Then the error should be nil", func() {
								So(err, ShouldBeNil)
								So(sols, ShouldNotBeNil)

								Convey("and the bindings should be as expected", func() {
									var got []types.TermResults
									for sols.Next() {
										m := types.TermResults{}
										err := sols.Scan(m)
										So(err, ShouldBeNil)

										got = append(got, m)
									}
									if tc.wantError != nil {
										So(sols.Err(), ShouldNotBeNil)
										So(sols.Err().Error(), ShouldEqual, tc.wantError.Error())
									} else {
										So(sols.Err(), ShouldBeNil)

										if tc.wantSuccess {
											So(len(got), ShouldBeGreaterThan, 0)
											So(len(got), ShouldEqual, len(tc.wantResult))
											for iGot, resultGot := range got {
												for varGot, termGot := range resultGot {
													So(testutil.ReindexUnknownVariables(termGot), ShouldEqual, tc.wantResult[iGot][varGot])
												}
											}
										} else {
											So(len(got), ShouldEqual, 0)
										}
									}
								})
							})
						})
					})
				})
			})
		}
	})
}

func TestEDDSAVerify(t *testing.T) {
	Convey("Given a test cases", t, func() {
		cases := []struct {
			program     string
			query       string
			wantResult  []types.TermResults
			wantError   error
			wantSuccess bool
		}{
			{ // All good
				program: `verify :-
hex_bytes('53167ac3fc4b720daa45b04fc73fe752578fa23a10048422d6904b7f4f7bba5a', PubKey),
hex_bytes('9b038f8ef6918cbb56040dfda401b56bb1ce79c472e7736e8677758c83367a9d', Msg),
hex_bytes('889bcfd331e8e43b5ebf430301dffb6ac9e2fce69f6227b43552fe3dc8cc1ee00c1cc53452a8712e9d5f80086dff8cf4999c1b93ed6c6e403c09334cb61ddd0b', Sig),
ecdsa_verify(PubKey, Msg, Sig, _).`,
				query:       `verify.`,
				wantResult:  []types.TermResults{{}},
				wantSuccess: true,
			},
			{ // Wrong Msg
				program: `verify :-
hex_bytes('53167ac3fc4b720daa45b04fc73fe752578fa23a10048422d6904b7f4f7bba5a', PubKey),
hex_bytes('9b038f8ef6918cbb56040dfda401b56bb1ce79c472e7736e8677758c83367a9e', Msg),
hex_bytes('889bcfd331e8e43b5ebf430301dffb6ac9e2fce69f6227b43552fe3dc8cc1ee00c1cc53452a8712e9d5f80086dff8cf4999c1b93ed6c6e403c09334cb61ddd0b', Sig),
ecdsa_verify(PubKey, Msg, Sig, _).`,
				query:       `verify.`,
				wantSuccess: false,
			},
			{ // Wrong public key
				program: `verify :-
hex_bytes('53167ac3fc4b720daa45b04fc73fe752578fa23a10048422d6904b7f4f7bba5b5b', PubKey),
hex_bytes('9b038f8ef6918cbb56040dfda401b56bb1ce79c472e7736e8677758c83367a9d', Msg),
hex_bytes('889bcfd331e8e43b5ebf430301dffb6ac9e2fce69f6227b43552fe3dc8cc1ee00c1cc53452a8712e9d5f80086dff8cf4999c1b93ed6c6e403c09334cb61ddd0b', Sig),
ecdsa_verify(PubKey, Msg, Sig, _).`,
				query:       `verify.`,
				wantSuccess: false,
				wantError:   fmt.Errorf("ecdsa_verify/4: failed verify signature: ed25519: bad public key length: 33"),
			},
			{ // Wrong signature
				program: `verify :-
hex_bytes('53167ac3fc4b720daa45b04fc73fe752578fa23a10048422d6904b7f4f7bba5a', PubKey),
hex_bytes('9b038f8ef6918cbb56040dfda401b56bb1ce79c472e7736e8677758c83367a9d', Msg),
hex_bytes('889bcfd331e8e43b5ebf430301dffb6ac9e2fce69f6227b43552fe3dc8cc1ee00c1cc53452a8712e9d5f80086dff', Sig),
ecdsa_verify(PubKey, Msg, Sig, _).`,
				query:       `verify.`,
				wantSuccess: false,
			},
			{
				// All good
				program: `verify :-
hex_bytes('2d2d2d2d2d424547494e205055424c4943204b45592d2d2d2d2d0d0a4d466b77457759484b6f5a497a6a3043415159494b6f5a497a6a3044415163445167414545386843612b527835565547393835506666565870433478446643660d0a6b75747a4c4b4d49586e6c383735734543525036654b4b79704c70514564564752526b3551396f687a64766b49392b583850756d666766356d673d3d0d0a2d2d2d2d2d454e44205055424c4943204b45592d2d2d2d2d', PubKey),
hex_bytes('e50c26e89f734b2ee12041ff27874c901891f74a0f0cf470333312a3034ce3be', Msg),
hex_bytes('30450220099e6f9dd218e0e304efa7a4224b0058a8e3aec73367ec239bee4ed8ed7d85db022100b504d3d0d2e879b04705c0e5a2b40b0521a5ab647ea207bd81134e1a4eb79e47', Sig),
secp_verify(PubKey, Msg, Sig, _).`,
				query:       `verify.`,
				wantResult:  []types.TermResults{{}},
				wantSuccess: true,
			},
			{ // Invalid secp signature
				program: `verify :-
hex_bytes('53167ac3fc4b720daa45b04fc73fe752578fa23a10048422d6904b7f4f7bba5a', PubKey),
hex_bytes('9b038f8ef6918cbb56040dfda401b56bb1ce79c472e7736e8677758c83367a9d', Msg),
hex_bytes('889bcfd331e8e43b5ebf430301dffb6ac9e2fce69f6227b43552fe3dc8cc1ee00c1cc53452a8712e9d5f80086dff8cf4999c1b93ed6c6e403c09334cb61ddd0b', Sig),
secp_verify(PubKey, Msg, Sig, _).`,
				query:       `verify.`,
				wantSuccess: false,
				wantError:   fmt.Errorf("secp_verify/4: failed verify signature: failed decode PEM public key"),
			},
			{
				// Wrong msg
				program: `verify :-
hex_bytes('2d2d2d2d2d424547494e205055424c4943204b45592d2d2d2d2d0d0a4d466b77457759484b6f5a497a6a3043415159494b6f5a497a6a3044415163445167414545386843612b527835565547393835506666565870433478446643660d0a6b75747a4c4b4d49586e6c383735734543525036654b4b79704c70514564564752526b3551396f687a64766b49392b583850756d666766356d673d3d0d0a2d2d2d2d2d454e44205055424c4943204b45592d2d2d2d2d', PubKey),
hex_bytes('e50c26e89f734b2ee12041ff27874c901891f74a0f0cf470333312a3034ce3bf', Msg),
hex_bytes('30450220099e6f9dd218e0e304efa7a4224b0058a8e3aec73367ec239bee4ed8ed7d85db022100b504d3d0d2e879b04705c0e5a2b40b0521a5ab647ea207bd81134e1a4eb79e47', Sig),
secp_verify(PubKey, Msg, Sig, _).`,
				query:       `verify.`,
				wantResult:  []types.TermResults{{}},
				wantSuccess: false,
			},
			{
				// Wrong signature
				program: `verify :-
hex_bytes('2d2d2d2d2d424547494e205055424c4943204b45592d2d2d2d2d0d0a4d466b77457759484b6f5a497a6a3043415159494b6f5a497a6a3044415163445167414545386843612b527835565547393835506666565870433478446643660d0a6b75747a4c4b4d49586e6c383735734543525036654b4b79704c70514564564752526b3551396f687a64766b49392b583850756d666766356d673d3d0d0a2d2d2d2d2d454e44205055424c4943204b45592d2d2d2d2d', PubKey),
hex_bytes('e50c26e89f734b2ee12041ff27874c901891f74a0f0cf470333312a3034ce3be', Msg),
hex_bytes('30450220099e6f9dd218e0e304efa7a4224b0058a8e3aec73367ec239bee4ed8ed7d85db022100b504d3d0d2e879b04705c0e5a2b40b0521a5ab647ea207bd81134e1a4eb79e48', Sig),
secp_verify(PubKey, Msg, Sig, _).`,
				query:       `verify.`,
				wantResult:  []types.TermResults{{}},
				wantSuccess: false,
			},
		}
		for nc, tc := range cases {
			Convey(fmt.Sprintf("Given the query #%d: %s", nc, tc.query), func() {
				Convey("and a context", func() {
					db := tmdb.NewMemDB()
					stateStore := store.NewCommitMultiStore(db)
					ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

					Convey("and a vm", func() {
						interpreter := testutil.NewLightInterpreterMust(ctx)
						interpreter.Register2(engine.NewAtom("hex_bytes"), HexBytes)
						interpreter.Register4(engine.NewAtom("ecdsa_verify"), ECDSAVerify)
						interpreter.Register4(engine.NewAtom("secp_verify"), SecpVerify)

						err := interpreter.Compile(ctx, tc.program)
						So(err, ShouldBeNil)

						Convey("When the predicate is called", func() {
							sols, err := interpreter.QueryContext(ctx, tc.query)

							Convey("Then the error should be nil", func() {
								So(err, ShouldBeNil)
								So(sols, ShouldNotBeNil)

								Convey("and the bindings should be as expected", func() {
									var got []types.TermResults
									for sols.Next() {
										m := types.TermResults{}
										err := sols.Scan(m)
										So(err, ShouldBeNil)

										got = append(got, m)
									}
									if tc.wantError != nil {
										So(sols.Err(), ShouldNotBeNil)
										So(sols.Err().Error(), ShouldEqual, tc.wantError.Error())
									} else {
										So(sols.Err(), ShouldBeNil)

										if tc.wantSuccess {
											So(len(got), ShouldBeGreaterThan, 0)
											So(len(got), ShouldEqual, len(tc.wantResult))
											for iGot, resultGot := range got {
												for varGot, termGot := range resultGot {
													So(testutil.ReindexUnknownVariables(termGot), ShouldEqual, tc.wantResult[iGot][varGot])
												}
											}
										} else {
											So(len(got), ShouldEqual, 0)
										}
									}
								})
							})
						})
					})
				})
			})
		}
	})
}
