// Copyright 2015 The go-Berith Authors
// This file is part of the go-Berith library.
//
// The go-Berith library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-Berith library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-Berith library. If not, see <http://www.gnu.org/licenses/>.

/*
[BERITH]
부트노드 에 대한 정보를 적는 곳
*/

package params

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main Berith network.
var MainnetBootnodes = []string{
	// Berith Foundation Go Bootnodes
	"enode://ea9b7c833a522780cb50dbb5f6e44c8d475ce8dedda44cb555e59994a5f89288908ebb288cfec9962c7321dee311a2a9bbfbad0178b1b3ef6dbcb33aea063e21@localhost:40404",
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{
	"enode://8142df22d4ca164db41f0cb6a66f439332ec8d9a799dbd1a5bbc990ca585fbc283dc949edb8a0962942accb0c361a1171dc16dd46abefd17b7dc8ef19e6894e8@localhost:55555",
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{}
