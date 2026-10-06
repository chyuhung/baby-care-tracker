// 构建后同步产物到 backend/dist（main.go `//go:embed dist` 需要它）。
// 全量清空再复制：vite 每次会重新输出所有 assets（含 hash 文件名），
// 残留旧文件只会让 embed 变大、且 go build 会把未删除的旧 hash 也打包进去。
import { rmSync, cpSync, mkdirSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..', '..')
const from = resolve(root, 'frontend', 'dist')
const to = resolve(root, 'backend', 'dist')

rmSync(to, { recursive: true, force: true })
mkdirSync(to, { recursive: true })
cpSync(from, to, { recursive: true })
process.stdout.write(`copied ${from} -> ${to}\n`)