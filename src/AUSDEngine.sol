//SPDX-License-Identifier: MIT

pragma solidity ^0.8.30;

contract AUSDEngine {
    error AUSDEngine__NotOwner();
    error AUSDEngine__AUSDAlreadyDefined();

    address private s_ausd;
    address private immutable i_owner;
    mapping(address user => mapping(address token => uint256 collateral)) private s_collateralDeposited;
    mapping(address user => uint256 debt) private s_totalDept;

    constructor() {
        i_owner = msg.sender;
    }

    modifier onlyOwner() {
        if(msg.sender != i_owner) {
            revert AUSDEngine__NotOwner();
        }
        _;
    }

    function setAUSD(address _ausd) public onlyOwner {
        if(s_ausd != address(0)) {
            revert AUSDEngine__AUSDAlreadyDefined();
        }

        s_ausd = _ausd;
    }
}