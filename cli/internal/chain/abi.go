package chain

// IdentityRegistryABI is the ABI for the IdentityRegistry contract.
const IdentityRegistryABI = `[
  {
    "inputs": [{"internalType": "bytes32", "name": "publicKey", "type": "bytes32"}],
    "name": "register",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [{"internalType": "address", "name": "user", "type": "address"}],
    "name": "getPublicKey",
    "outputs": [{"internalType": "bytes32", "name": "", "type": "bytes32"}],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [{"internalType": "address", "name": "user", "type": "address"}],
    "name": "isRegistered",
    "outputs": [{"internalType": "bool", "name": "", "type": "bool"}],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "anonymous": false,
    "inputs": [
      {"indexed": true, "internalType": "address", "name": "user", "type": "address"},
      {"indexed": false, "internalType": "bytes32", "name": "publicKey", "type": "bytes32"}
    ],
    "name": "Registered",
    "type": "event"
  }
]`

// ProjectVaultABI is the ABI for the ProjectVault contract.
const ProjectVaultABI = `[
  {
    "inputs": [{"internalType": "address", "name": "registryAddress", "type": "address"}],
    "stateMutability": "nonpayable",
    "type": "constructor"
  },
  {
    "inputs": [{"internalType": "string", "name": "name", "type": "string"}],
    "name": "createProject",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {"internalType": "bytes32", "name": "projectId", "type": "bytes32"},
      {"internalType": "address", "name": "member", "type": "address"}
    ],
    "name": "addMember",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {"internalType": "bytes32", "name": "projectId", "type": "bytes32"},
      {"internalType": "string", "name": "name", "type": "string"},
      {"internalType": "address[]", "name": "members", "type": "address[]"},
      {"internalType": "bytes[]", "name": "encryptedValues", "type": "bytes[]"}
    ],
    "name": "setSecret",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {"internalType": "bytes32", "name": "projectId", "type": "bytes32"},
      {"internalType": "string", "name": "name", "type": "string"}
    ],
    "name": "getSecret",
    "outputs": [{"internalType": "bytes", "name": "", "type": "bytes"}],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [{"internalType": "bytes32", "name": "projectId", "type": "bytes32"}],
    "name": "getMembers",
    "outputs": [{"internalType": "address[]", "name": "", "type": "address[]"}],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [{"internalType": "bytes32", "name": "projectId", "type": "bytes32"}],
    "name": "getSecretNames",
    "outputs": [{"internalType": "string[]", "name": "", "type": "string[]"}],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [{"internalType": "bytes32", "name": "projectId", "type": "bytes32"}],
    "name": "getOwner",
    "outputs": [{"internalType": "address", "name": "", "type": "address"}],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [{"internalType": "address", "name": "member", "type": "address"}],
    "name": "getMemberProjects",
    "outputs": [{"internalType": "bytes32[]", "name": "", "type": "bytes32[]"}],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {"internalType": "bytes32", "name": "projectId", "type": "bytes32"},
      {"internalType": "address", "name": "user", "type": "address"}
    ],
    "name": "isMember",
    "outputs": [{"internalType": "bool", "name": "", "type": "bool"}],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [{"internalType": "bytes32", "name": "", "type": "bytes32"}],
    "name": "projectNames",
    "outputs": [{"internalType": "string", "name": "", "type": "string"}],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "anonymous": false,
    "inputs": [
      {"indexed": true, "internalType": "bytes32", "name": "projectId", "type": "bytes32"},
      {"indexed": true, "internalType": "address", "name": "owner", "type": "address"},
      {"indexed": false, "internalType": "string", "name": "name", "type": "string"}
    ],
    "name": "ProjectCreated",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {"indexed": true, "internalType": "bytes32", "name": "projectId", "type": "bytes32"},
      {"indexed": true, "internalType": "address", "name": "member", "type": "address"}
    ],
    "name": "MemberAdded",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {"indexed": true, "internalType": "bytes32", "name": "projectId", "type": "bytes32"},
      {"indexed": false, "internalType": "string", "name": "name", "type": "string"}
    ],
    "name": "SecretSet",
    "type": "event"
  }
]`
