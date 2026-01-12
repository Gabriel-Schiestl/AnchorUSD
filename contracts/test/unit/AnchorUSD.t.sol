//SPDX-License-Identifier: MIT

pragma solidity ^0.8.30;

import {Test} from "forge-std/Test.sol";
import {AnchorUSD} from "../../src/AnchorUSD.sol";
import {AUSDEngine} from "../../src/AUSDEngine.sol";
import {DeployAUSD} from "../../script/DeployAUSD.s.sol";

contract AnchorUSDTest is Test {
    AnchorUSD aUSD;
    AUSDEngine engine;
    DeployAUSD deployer;

    address private USER = makeAddr("user");
    uint256 private constant INITIAL_BALANCE = 10;
    uint256 private constant BURN_AMOUNT = 5;

    function setUp() external {
        deployer = new DeployAUSD();
        (aUSD, engine, ) = deployer.run();
    }

    function testRevertIfNotEngine() external {
        vm.expectRevert(AnchorUSD.AnchorUSD__OnlyEngine.selector);
        aUSD.mint(USER, INITIAL_BALANCE);

        vm.expectRevert(AnchorUSD.AnchorUSD__OnlyEngine.selector);
        aUSD.burn(USER, INITIAL_BALANCE);
    }

    modifier mint() {
        vm.prank(address(engine));
        aUSD.mint(USER, INITIAL_BALANCE);
        _;
    }

    function testMint() external mint {
        assertEq(aUSD.balanceOf(USER), INITIAL_BALANCE);
        assertEq(aUSD.totalSupply(), aUSD.balanceOf(USER));
    }

    function testBurn() external mint {
        vm.prank(address(engine));
        aUSD.burn(USER, BURN_AMOUNT);

        assertEq(aUSD.balanceOf(USER), BURN_AMOUNT);
        assertEq(aUSD.totalSupply(), aUSD.balanceOf(USER));
    }

    function testRevertBurnIfNoBalance() external {
        vm.expectRevert();
        aUSD.burn(USER, BURN_AMOUNT);
    }

    function testEngineAddress() external view {
        assertEq(aUSD.i_engine(), address(engine));
    }
}
