import type { NavigateToNextCellParams } from 'ag-grid-community/dist/lib/interfaces/iCallbackParams'

export function selectEntireRow(params: NavigateToNextCellParams) {
  const suggestedNextCell = params.nextCellPosition

  // this is some code
  const KEY_UP = 'ArrowUp'
  const KEY_DOWN = 'ArrowDown'

  const noUpOrDownKey = params.key !== KEY_DOWN && params.key !== KEY_UP
  if (noUpOrDownKey) {
    return suggestedNextCell
  }

  params.api.forEachNode((node: any) => {
    if (!node || !suggestedNextCell) return
    if (node.rowIndex === suggestedNextCell.rowIndex) {
      node.setSelected(true)
    }
  })

  return suggestedNextCell
}
