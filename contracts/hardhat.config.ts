import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import * as fs from "fs";
import * as os from "os";
import * as path from "path";

function loadEnparse(): { rpcUrl: string; privateKey: string } {
  const dir = path.join(os.homedir(), ".enparse");
  let rpcUrl = "http://127.0.0.1:8545";
  let privateKey = "";

  try {
    const cfg = JSON.parse(fs.readFileSync(path.join(dir, "config.json"), "utf8"));
    if (cfg.rpc_url) rpcUrl = cfg.rpc_url;
  } catch {}

  try {
    const id = JSON.parse(fs.readFileSync(path.join(dir, "identity.json"), "utf8"));
    if (id.eth_privkey) privateKey = "0x" + id.eth_privkey;
  } catch {}

  return { rpcUrl, privateKey };
}

const { rpcUrl, privateKey } = loadEnparse();

const config: HardhatUserConfig = {
  solidity: {
    version: "0.8.24",
    settings: {
      evmVersion: "cancun",
    },
  },
  networks: {
    hardhat: {
      hardfork: "cancun",
    },
    sepolia: {
      url: rpcUrl,
      accounts: privateKey ? [privateKey] : [],
      chainId: 11155111,
    },
  },
};

export default config;
