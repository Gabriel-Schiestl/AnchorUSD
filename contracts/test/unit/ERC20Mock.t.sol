// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

import {Test} from "forge-std/Test.sol";
import {ERC20Mock} from "../mocks/ERC20Mock.sol";

contract ERC20MockTest is Test {
    ERC20Mock token;

    address OWNER = makeAddr("owner");
    address ALICE = makeAddr("alice");
    address BOB = makeAddr("bob");

    string constant TOKEN_NAME = "Mock Token";
    string constant TOKEN_SYMBOL = "MTK";
    uint256 constant INITIAL_SUPPLY = 1000 ether;
    uint256 constant MINT_AMOUNT = 100 ether;
    uint256 constant BURN_AMOUNT = 50 ether;
    uint256 constant TRANSFER_AMOUNT = 200 ether;
    uint256 constant APPROVE_AMOUNT = 500 ether;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(
        address indexed owner,
        address indexed spender,
        uint256 value
    );

    function setUp() external {
        token = new ERC20Mock(TOKEN_NAME, TOKEN_SYMBOL, OWNER, INITIAL_SUPPLY);
    }

    //////// Constructor Tests ////////

    function testConstructorSetsNameAndSymbol() external view {
        assertEq(token.name(), TOKEN_NAME);
        assertEq(token.symbol(), TOKEN_SYMBOL);
    }

    function testConstructorMintsInitialSupply() external view {
        assertEq(token.balanceOf(OWNER), INITIAL_SUPPLY);
        assertEq(token.totalSupply(), INITIAL_SUPPLY);
    }

    function testConstructorWithZeroInitialSupply() external {
        ERC20Mock newToken = new ERC20Mock(TOKEN_NAME, TOKEN_SYMBOL, ALICE, 0);

        assertEq(newToken.balanceOf(ALICE), 0);
        assertEq(newToken.totalSupply(), 0);
    }

    function testConstructorWithZeroAddress() external {
        vm.expectRevert();
        new ERC20Mock(TOKEN_NAME, TOKEN_SYMBOL, address(0), INITIAL_SUPPLY);
    }

    //////// Mint Tests ////////

    function testMintIncreasesBalance() external {
        uint256 balanceBefore = token.balanceOf(ALICE);

        token.mint(ALICE, MINT_AMOUNT);

        assertEq(token.balanceOf(ALICE), balanceBefore + MINT_AMOUNT);
    }

    function testMintIncreasesTotalSupply() external {
        uint256 totalSupplyBefore = token.totalSupply();

        token.mint(ALICE, MINT_AMOUNT);

        assertEq(token.totalSupply(), totalSupplyBefore + MINT_AMOUNT);
    }

    function testMintEmitsTransferEvent() external {
        vm.expectEmit(true, true, false, true, address(token));
        emit Transfer(address(0), ALICE, MINT_AMOUNT);

        token.mint(ALICE, MINT_AMOUNT);
    }

    function testMintToZeroAddressReverts() external {
        vm.expectRevert();
        token.mint(address(0), MINT_AMOUNT);
    }

    function testMintZeroAmount() external {
        uint256 balanceBefore = token.balanceOf(ALICE);

        token.mint(ALICE, 0);

        assertEq(token.balanceOf(ALICE), balanceBefore);
    }

    function testMultipleMints() external {
        token.mint(ALICE, MINT_AMOUNT);
        token.mint(ALICE, MINT_AMOUNT);

        assertEq(token.balanceOf(ALICE), MINT_AMOUNT * 2);
    }

    //////// Burn Tests ////////

    function testBurnDecreasesBalance() external {
        token.mint(ALICE, MINT_AMOUNT);

        token.burn(ALICE, BURN_AMOUNT);

        assertEq(token.balanceOf(ALICE), MINT_AMOUNT - BURN_AMOUNT);
    }

    function testBurnDecreasesTotalSupply() external {
        uint256 totalSupplyBefore = token.totalSupply();

        token.burn(OWNER, BURN_AMOUNT);

        assertEq(token.totalSupply(), totalSupplyBefore - BURN_AMOUNT);
    }

    function testBurnEmitsTransferEvent() external {
        vm.expectEmit(true, true, false, true, address(token));
        emit Transfer(OWNER, address(0), BURN_AMOUNT);

        token.burn(OWNER, BURN_AMOUNT);
    }

    function testBurnMoreThanBalanceReverts() external {
        vm.expectRevert();
        token.burn(ALICE, 1 ether);
    }

    function testBurnFromZeroAddressReverts() external {
        vm.expectRevert();
        token.burn(address(0), BURN_AMOUNT);
    }

    function testBurnZeroAmount() external {
        uint256 balanceBefore = token.balanceOf(OWNER);

        token.burn(OWNER, 0);

        assertEq(token.balanceOf(OWNER), balanceBefore);
    }

    function testBurnEntireBalance() external {
        token.mint(ALICE, MINT_AMOUNT);

        token.burn(ALICE, MINT_AMOUNT);

        assertEq(token.balanceOf(ALICE), 0);
    }

    //////// TransferInternal Tests ////////

    function testTransferInternalMovesTokens() external {
        token.transferInternal(OWNER, ALICE, TRANSFER_AMOUNT);

        assertEq(token.balanceOf(OWNER), INITIAL_SUPPLY - TRANSFER_AMOUNT);
        assertEq(token.balanceOf(ALICE), TRANSFER_AMOUNT);
    }

    function testTransferInternalDoesNotChangeTotalSupply() external {
        uint256 totalSupplyBefore = token.totalSupply();

        token.transferInternal(OWNER, ALICE, TRANSFER_AMOUNT);

        assertEq(token.totalSupply(), totalSupplyBefore);
    }

    function testTransferInternalEmitsTransferEvent() external {
        vm.expectEmit(true, true, false, true, address(token));
        emit Transfer(OWNER, ALICE, TRANSFER_AMOUNT);

        token.transferInternal(OWNER, ALICE, TRANSFER_AMOUNT);
    }

    function testTransferInternalFromZeroAddressReverts() external {
        vm.expectRevert();
        token.transferInternal(address(0), ALICE, TRANSFER_AMOUNT);
    }

    function testTransferInternalToZeroAddressReverts() external {
        vm.expectRevert();
        token.transferInternal(OWNER, address(0), TRANSFER_AMOUNT);
    }

    function testTransferInternalInsufficientBalanceReverts() external {
        vm.expectRevert();
        token.transferInternal(ALICE, BOB, 1 ether);
    }

    function testTransferInternalZeroAmount() external {
        uint256 ownerBalanceBefore = token.balanceOf(OWNER);
        uint256 aliceBalanceBefore = token.balanceOf(ALICE);

        token.transferInternal(OWNER, ALICE, 0);

        assertEq(token.balanceOf(OWNER), ownerBalanceBefore);
        assertEq(token.balanceOf(ALICE), aliceBalanceBefore);
    }

    function testTransferInternalToSelf() external {
        uint256 balanceBefore = token.balanceOf(OWNER);

        token.transferInternal(OWNER, OWNER, TRANSFER_AMOUNT);

        assertEq(token.balanceOf(OWNER), balanceBefore);
    }

    //////// ApproveInternal Tests ////////

    function testApproveInternalSetsAllowance() external {
        token.approveInternal(OWNER, ALICE, APPROVE_AMOUNT);

        assertEq(token.allowance(OWNER, ALICE), APPROVE_AMOUNT);
    }

    function testApproveInternalEmitsApprovalEvent() external {
        vm.expectEmit(true, true, false, true, address(token));
        emit Approval(OWNER, ALICE, APPROVE_AMOUNT);

        token.approveInternal(OWNER, ALICE, APPROVE_AMOUNT);
    }

    function testApproveInternalOwnerZeroAddressReverts() external {
        vm.expectRevert();
        token.approveInternal(address(0), ALICE, APPROVE_AMOUNT);
    }

    function testApproveInternalSpenderZeroAddressReverts() external {
        vm.expectRevert();
        token.approveInternal(OWNER, address(0), APPROVE_AMOUNT);
    }

    function testApproveInternalZeroAmount() external {
        token.approveInternal(OWNER, ALICE, 0);

        assertEq(token.allowance(OWNER, ALICE), 0);
    }

    function testApproveInternalOverwritesPreviousAllowance() external {
        token.approveInternal(OWNER, ALICE, APPROVE_AMOUNT);
        token.approveInternal(OWNER, ALICE, MINT_AMOUNT);

        assertEq(token.allowance(OWNER, ALICE), MINT_AMOUNT);
    }

    //////// Standard ERC20 Functions Tests ////////

    function testStandardTransfer() external {
        vm.prank(OWNER);
        bool success = token.transfer(ALICE, TRANSFER_AMOUNT);

        assertTrue(success);
        assertEq(token.balanceOf(OWNER), INITIAL_SUPPLY - TRANSFER_AMOUNT);
        assertEq(token.balanceOf(ALICE), TRANSFER_AMOUNT);
    }

    function testStandardApprove() external {
        vm.prank(OWNER);
        bool success = token.approve(ALICE, APPROVE_AMOUNT);

        assertTrue(success);
        assertEq(token.allowance(OWNER, ALICE), APPROVE_AMOUNT);
    }

    function testStandardTransferFrom() external {
        vm.prank(OWNER);
        token.approve(ALICE, APPROVE_AMOUNT);

        vm.prank(ALICE);
        bool success = token.transferFrom(OWNER, BOB, TRANSFER_AMOUNT);

        assertTrue(success);
        assertEq(token.balanceOf(OWNER), INITIAL_SUPPLY - TRANSFER_AMOUNT);
        assertEq(token.balanceOf(BOB), TRANSFER_AMOUNT);
        assertEq(
            token.allowance(OWNER, ALICE),
            APPROVE_AMOUNT - TRANSFER_AMOUNT
        );
    }

    //////// Integration Tests ////////

    function testMintAndBurn() external {
        token.mint(ALICE, MINT_AMOUNT);
        assertEq(token.balanceOf(ALICE), MINT_AMOUNT);

        token.burn(ALICE, BURN_AMOUNT);
        assertEq(token.balanceOf(ALICE), MINT_AMOUNT - BURN_AMOUNT);
    }

    function testMintTransferAndBurn() external {
        token.mint(ALICE, MINT_AMOUNT);
        token.transferInternal(ALICE, BOB, BURN_AMOUNT);
        token.burn(BOB, BURN_AMOUNT);

        assertEq(token.balanceOf(ALICE), MINT_AMOUNT - BURN_AMOUNT);
        assertEq(token.balanceOf(BOB), 0);
    }

    function testComplexScenario() external {
        // Mint to Alice
        token.mint(ALICE, MINT_AMOUNT);

        // Alice approves Bob
        token.approveInternal(ALICE, BOB, APPROVE_AMOUNT);

        // Transfer from Owner to Alice
        token.transferInternal(OWNER, ALICE, TRANSFER_AMOUNT);

        // Burn some of Alice's tokens
        token.burn(ALICE, BURN_AMOUNT);

        // Verify final state
        assertEq(
            token.balanceOf(ALICE),
            MINT_AMOUNT + TRANSFER_AMOUNT - BURN_AMOUNT
        );
        assertEq(token.allowance(ALICE, BOB), APPROVE_AMOUNT);
        assertEq(
            token.totalSupply(),
            INITIAL_SUPPLY + MINT_AMOUNT - BURN_AMOUNT
        );
    }

    //////// Edge Cases ////////

    function testMaxUint256Mint() external {
        // This should revert due to overflow protection
        vm.expectRevert();
        token.mint(ALICE, type(uint256).max);
    }

    function testMultipleApprovals() external {
        token.approveInternal(OWNER, ALICE, 100 ether);
        token.approveInternal(OWNER, BOB, 200 ether);
        token.approveInternal(OWNER, ALICE, 300 ether);

        assertEq(token.allowance(OWNER, ALICE), 300 ether);
        assertEq(token.allowance(OWNER, BOB), 200 ether);
    }

    function testTransferInternalChain() external {
        token.transferInternal(OWNER, ALICE, 100 ether);
        token.transferInternal(ALICE, BOB, 50 ether);
        token.transferInternal(BOB, OWNER, 25 ether);

        assertEq(token.balanceOf(OWNER), INITIAL_SUPPLY - 75 ether);
        assertEq(token.balanceOf(ALICE), 50 ether);
        assertEq(token.balanceOf(BOB), 25 ether);
    }

    //////// Fuzz Tests ////////

    function testFuzzMint(address to, uint256 amount) external {
        vm.assume(to != address(0));
        vm.assume(amount < type(uint256).max - token.totalSupply());

        uint256 balanceBefore = token.balanceOf(to);
        uint256 totalSupplyBefore = token.totalSupply();

        token.mint(to, amount);

        assertEq(token.balanceOf(to), balanceBefore + amount);
        assertEq(token.totalSupply(), totalSupplyBefore + amount);
    }

    function testFuzzBurn(
        address from,
        uint256 mintAmount,
        uint256 burnAmount
    ) external {
        vm.assume(from != address(0));
        vm.assume(mintAmount < type(uint256).max / 2);
        vm.assume(burnAmount <= mintAmount);

        uint256 balanceBefore = token.balanceOf(from);

        token.mint(from, mintAmount);
        token.burn(from, burnAmount);

        assertEq(
            token.balanceOf(from),
            balanceBefore + mintAmount - burnAmount
        );
    }

    function testFuzzTransferInternal(uint256 amount) external {
        vm.assume(amount <= INITIAL_SUPPLY);

        token.transferInternal(OWNER, ALICE, amount);

        assertEq(token.balanceOf(OWNER), INITIAL_SUPPLY - amount);
        assertEq(token.balanceOf(ALICE), amount);
    }

    function testFuzzApproveInternal(
        address owner,
        address spender,
        uint256 amount
    ) external {
        vm.assume(owner != address(0));
        vm.assume(spender != address(0));

        token.approveInternal(owner, spender, amount);

        assertEq(token.allowance(owner, spender), amount);
    }
}
