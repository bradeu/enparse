// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./IdentityRegistry.sol";

contract ProjectVault {
    IdentityRegistry public immutable registry;

    struct Project {
        address owner;
        address[] members;
        bool exists;
    }

    mapping(bytes32 => Project) private projects;
    // projectId => secretName => member => encryptedBlob
    mapping(bytes32 => mapping(string => mapping(address => bytes))) private secrets;
    mapping(bytes32 => string[]) private secretNameList;
    mapping(bytes32 => mapping(string => bool)) private secretAdded;
    mapping(address => bytes32[]) private memberProjects;
    mapping(bytes32 => string) public projectNames;

    event ProjectCreated(bytes32 indexed projectId, address indexed owner, string name);
    event MemberAdded(bytes32 indexed projectId, address indexed member);
    event SecretSet(bytes32 indexed projectId, string name);

    constructor(address registryAddress) {
        registry = IdentityRegistry(registryAddress);
    }

    modifier onlyOwner(bytes32 projectId) {
        require(projects[projectId].owner == msg.sender, "Not project owner");
        _;
    }

    modifier projectExists(bytes32 projectId) {
        require(projects[projectId].exists, "Project does not exist");
        _;
    }

    function createProject(string calldata name) external {
        require(bytes(name).length > 0, "Name required");
        bytes32 projectId = keccak256(bytes(name));
        require(!projects[projectId].exists, "Project already exists");
        require(registry.isRegistered(msg.sender), "Must register identity first");

        Project storage p = projects[projectId];
        p.owner = msg.sender;
        p.exists = true;
        p.members.push(msg.sender);
        memberProjects[msg.sender].push(projectId);
        projectNames[projectId] = name;

        emit ProjectCreated(projectId, msg.sender, name);
    }

    function addMember(bytes32 projectId, address member)
        external
        projectExists(projectId)
        onlyOwner(projectId)
    {
        require(registry.isRegistered(member), "Member must register identity first");
        require(!isMember(projectId, member), "Already a member");

        projects[projectId].members.push(member);
        memberProjects[member].push(projectId);

        emit MemberAdded(projectId, member);
    }

    function setSecret(
        bytes32 projectId,
        string calldata name,
        address[] calldata members,
        bytes[] calldata encryptedValues
    ) external projectExists(projectId) onlyOwner(projectId) {
        require(members.length == encryptedValues.length, "Length mismatch");
        require(members.length > 0, "No members provided");

        for (uint256 i = 0; i < members.length; i++) {
            secrets[projectId][name][members[i]] = encryptedValues[i];
        }
        if (!secretAdded[projectId][name]) {
            secretNameList[projectId].push(name);
            secretAdded[projectId][name] = true;
        }

        emit SecretSet(projectId, name);
    }

    function getSecret(bytes32 projectId, string calldata name)
        external
        view
        projectExists(projectId)
        returns (bytes memory)
    {
        require(isMember(projectId, msg.sender), "Not a project member");
        return secrets[projectId][name][msg.sender];
    }

    function getMembers(bytes32 projectId)
        external
        view
        projectExists(projectId)
        returns (address[] memory)
    {
        return projects[projectId].members;
    }

    function getSecretNames(bytes32 projectId)
        external
        view
        projectExists(projectId)
        returns (string[] memory)
    {
        require(isMember(projectId, msg.sender), "Not a project member");
        return secretNameList[projectId];
    }

    function getOwner(bytes32 projectId)
        external
        view
        projectExists(projectId)
        returns (address)
    {
        return projects[projectId].owner;
    }

    function getMemberProjects(address member) external view returns (bytes32[] memory) {
        return memberProjects[member];
    }

    function isMember(bytes32 projectId, address user) public view returns (bool) {
        address[] storage members = projects[projectId].members;
        for (uint256 i = 0; i < members.length; i++) {
            if (members[i] == user) return true;
        }
        return false;
    }
}
