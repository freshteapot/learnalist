import fs from 'fs';
import chokidar from 'chokidar';

const chromeWatch = (cwd, files) => {
    const watchFiles = Object.values(files).map(obj => obj.from);

    const watcher = chokidar.watch(watchFiles, {
        ignored: /(^|[\/\\])\../, // ignore dotfiles
        persistent: true
    });

    // Something to use when events are received.
    const log = console.log.bind(console);
    // Add event listeners.
    watcher
        .on('add', path => sync(files))
        .on('change', path => sync(files))
        .on('unlink', path => sync(files));

}

const sync = (files) => {
    Object.values(files).map(obj => {
        fs.copyFileSync(obj.from, obj.to);
    })
}

export { chromeWatch }
