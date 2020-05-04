import { push } from 'svelte-spa-router';

let paths = {
  list: {
    edit: (uuid) => {
      push("/list/edit/" + uuid);
    },
    view: (uuid) => {
      push("/list/view/" + uuid);
    }
  }
}
export default paths;
