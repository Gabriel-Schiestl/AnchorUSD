//SPDX-License-Identifier: MIT

pragma solidity ^0.8.30;

import {ERC20Burnable, ERC20} from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";

contract AnchorUSD is ERC20 {
    address public immutable i_engine;

    error AnchorUSD__OnlyEngine();

    constructor(address _engine) ERC20("Anchor USD", "aUSD") {
        i_engine = _engine;
    }

    modifier onlyEngine() {
        if (msg.sender != i_engine) revert AnchorUSD__OnlyEngine();
        _;
    }

    function mint(address to, uint256 amount) external onlyEngine {
        _mint(to, amount);
    }

    function burn(address from, uint256 amount) external onlyEngine {
        _burn(from, amount);
    }
}
