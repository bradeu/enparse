// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

contract IdentityRegistry {
    mapping(address => bytes32) private publicKeys;

    event Registered(address indexed user, bytes32 publicKey);

    function register(bytes32 publicKey) external {
        require(publicKey != bytes32(0), "Invalid public key");
        publicKeys[msg.sender] = publicKey;
        emit Registered(msg.sender, publicKey);
    }

    function getPublicKey(address user) external view returns (bytes32) {
        return publicKeys[user];
    }

    function isRegistered(address user) external view returns (bool) {
        return publicKeys[user] != bytes32(0);
    }
}
