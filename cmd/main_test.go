package main

import (
	"testing"
)

func TestValidateBoardSize(t *testing.T) {
	tests := []struct {
		width, height int
		expected      bool
	}{
		{5, 5, false},
		{6, 6, true},
		{12, 12, true},
		{13, 10, false},
		{10, 13, false},
	}

	for _, test := range tests {
		result := validateBoardSize(test.width) && validateBoardSize(test.height)
		if result != test.expected {
			t.Errorf("validateBoardSize(%d, %d) = %v; want %v\n", test.width, test.height, result, test.expected)
		}
	}
}

func TestValidateCoordinate(t *testing.T) {
	model := Model{
		Board: Board{Width: 8, Height: 8},
	}

	tests := []struct {
		coord    string
		expected bool
	}{
		{"A1", true},
		{"H8", true},
		{"I8", false},  // Column out of range
		{"A9", false},  // Row out of range
		{"A0", false},  // Invalid row
		{"J1", false},  // Invalid column
		{"AA", false},  // Invalid format
		{"11", false},  // Invalid format
		{"", false},    // Empty string
		{"A1B", false}, // Too long
	}

	for _, test := range tests {
		result := validateCoordinate(test.coord, model)
		if result != test.expected {
			t.Errorf("validateCoordinate(%q) = %v; want %v", test.coord, result, test.expected)
		}
	}
}

func TestPieceIdentification(t *testing.T) {
	tests := []struct {
		piece       rune
		isWhite     bool
		isBlack     bool
		description string
	}{
		{WhiteKing, true, false, "White King"},
		{WhiteTower, true, false, "White Tower"},
		{WhiteHorse, true, false, "White Horse"},
		{BlackKing, false, true, "Black King"},
		{BlackTower, false, true, "Black Tower"},
		{BlackHorse, false, true, "Black Horse"},
		{EC, false, false, "Empty Cell"},
		{'X', false, false, "Unknown piece"},
	}

	for _, test := range tests {
		if got := isWhitePiece(test.piece); got != test.isWhite {
			t.Errorf("isWhitePiece(%s) = %v; want %v", test.description, got, test.isWhite)
		}
		if got := isBlackPiece(test.piece); got != test.isBlack {
			t.Errorf("isBlackPiece(%s) = %v; want %v", test.description, got, test.isBlack)
		}
	}
}

func TestValidMoves(t *testing.T) {
	tests := []struct {
		name      string
		moveFunc  func(fromCol, fromRow, toCol, toRow int) bool
		testCases []struct {
			fromCol, fromRow, toCol, toRow int
			expected                       bool
		}
	}{
		{
			name:     "King Movement",
			moveFunc: isValidKingMove,
			testCases: []struct {
				fromCol, fromRow, toCol, toRow int
				expected                       bool
			}{
				{4, 4, 4, 4, false}, // Same position
				{4, 4, 4, 5, true},  // Up
				{4, 4, 5, 5, true},  // Diagonal
				{4, 4, 6, 6, false}, // Too far
				{4, 4, 4, 6, false}, // Too far straight
			},
		},
		{
			name:     "Tower Movement",
			moveFunc: isValidTowerMove,
			testCases: []struct {
				fromCol, fromRow, toCol, toRow int
				expected                       bool
			}{
				{4, 4, 4, 4, false}, // Same position
				{4, 4, 4, 7, true},  // Vertical within 3
				{4, 4, 7, 4, true},  // Horizontal within 3
				{4, 4, 6, 6, true},  // Diagonal within 3
				{4, 4, 8, 4, false}, // Too far horizontal
				{4, 4, 4, 8, false}, // Too far vertical
				{4, 4, 7, 6, false}, // Invalid pattern
			},
		},
		{
			name:     "Horse Movement",
			moveFunc: isValidHorseMove,
			testCases: []struct {
				fromCol, fromRow, toCol, toRow int
				expected                       bool
			}{
				{4, 4, 4, 4, false}, // Same position
				{4, 4, 6, 5, true},  // L shape
				{4, 4, 5, 6, true},  // L shape other direction
				{4, 4, 6, 6, false}, // Diagonal
				{4, 4, 4, 5, false}, // Straight
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, tc := range test.testCases {
				got := test.moveFunc(tc.fromCol, tc.fromRow, tc.toCol, tc.toRow)
				if got != tc.expected {
					t.Errorf("Move from (%d,%d) to (%d,%d) = %v; want %v",
						tc.fromCol, tc.fromRow, tc.toCol, tc.toRow, got, tc.expected)
				}
			}
		})
	}
}
