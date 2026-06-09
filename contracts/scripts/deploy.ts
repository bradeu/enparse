import { ethers } from "hardhat";
import * as fs from "fs";
import * as os from "os";
import * as path from "path";

async function main(): Promise<void> {
  const [deployer] = await ethers.getSigners();
  console.log("Deployer:", deployer.address);
  console.log(
    "Balance: ",
    ethers.formatEther(await ethers.provider.getBalance(deployer.address)),
    "ETH"
  );

  const IdentityRegistry = await ethers.getContractFactory("IdentityRegistry");
  const registry = await IdentityRegistry.deploy();
  await registry.waitForDeployment();
  const registryAddr = await registry.getAddress();
  console.log("IdentityRegistry:  ", registryAddr);

  const ProjectVault = await ethers.getContractFactory("ProjectVault");
  const vault = await ProjectVault.deploy(registryAddr);
  await vault.waitForDeployment();
  const vaultAddr = await vault.getAddress();
  console.log("ProjectVault:      ", vaultAddr);

  // Save deploy receipt alongside contracts/
  const outPath = path.join(__dirname, "..", "deploy.sepolia.json");
  fs.writeFileSync(outPath, JSON.stringify({
    network: "sepolia",
    chain_id: 11155111,
    identity_registry_addr: registryAddr,
    project_vault_addr: vaultAddr,
    deployer: deployer.address,
    deployed_at: new Date().toISOString(),
  }, null, 2));
  console.log("Saved → deploy.sepolia.json");

  // Write addresses directly into ~/.enparse/config.json
  const cfgPath = path.join(os.homedir(), ".enparse", "config.json");
  let cfg: Record<string, string> = {};
  try {
    cfg = JSON.parse(fs.readFileSync(cfgPath, "utf8"));
  } catch {}
  cfg.identity_registry_addr = registryAddr;
  cfg.project_vault_addr = vaultAddr;
  fs.mkdirSync(path.dirname(cfgPath), { recursive: true });
  fs.writeFileSync(cfgPath, JSON.stringify(cfg, null, 2), { mode: 0o600 });
  console.log("Updated ~/.enparse/config.json with contract addresses");
}

main().catch((err: Error) => {
  console.error(err);
  process.exit(1);
});
