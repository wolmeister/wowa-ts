import { format } from 'date-fns';
import type { LocalAddon } from './addon.repository';

export class AddonPrinter {
  print(addons: LocalAddon[]): void {
    const table = [['ID', 'Name', 'Version', 'Game Version', 'Updated at']];
    const largestColumns = table[0].map((c) => c.length);
    for (const addon of addons) {
      const addonName = addon.name.split(' ... ')[0];
      const updatedAt = addon.updatedAt ? format(addon.updatedAt, 'dd/MM/yyyy HH:mm') : 'N/A';

      // TODO: Improve this.
      table.push([addon.id, addonName, addon.version, addon.gameVersion, updatedAt]);
      largestColumns[0] = Math.max(addon.id.length, largestColumns[0]);
      largestColumns[1] = Math.max(addonName.length, largestColumns[1]);
      largestColumns[2] = Math.max(addon.version.length, largestColumns[2]);
      largestColumns[3] = Math.max(addon.gameVersion.length, largestColumns[3]);
      largestColumns[4] = Math.max(updatedAt.length, largestColumns[4]);
    }

    for (const largestColumn of largestColumns) {
      for (let i = 0; i < largestColumn + 2; i++) {
        console.write('-');
      }
    }

    console.write('-\n');

    for (const row of table) {
      for (let columnIndex = 0; columnIndex < row.length; columnIndex++) {
        const column = row[columnIndex];
        console.write('| ');
        console.write(column);

        for (let i = column.length; i < largestColumns[columnIndex]; i++) {
          console.write(' ');
        }

        if (columnIndex === row.length - 1) {
          console.write('|');
        }
      }

      console.write('\n');
    }

    for (const largestColumn of largestColumns) {
      for (let i = 0; i < largestColumn + 2; i++) {
        console.write('-');
      }
    }

    console.write('-\n');
  }
}
