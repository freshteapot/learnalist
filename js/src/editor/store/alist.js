import { writable } from 'svelte/store';


const {subscribe, set, update} = writable({
  data: [],
  info: {
    title: title,
    type: listType,
    labels: [],
    sharedWith: 'private'
  }
});

const loading = writable(false);
const error = writable('');

const AlistStore = () => ({
    subscribe,
    set,
    loading,
    error,
    async create(query) {
      try {
          error.set('');
          loading.set(true);
          const response = await axios.get(`${BASE_URL}/search.json?title=${query}`);
          loading.set(false);
          set(response.data.docs);
          return response.data;
      } catch(e) {
          loading.set(false);
          set([]);
          error.set(`Error has been occurred. Details: ${e.message}`);
      }
    },

    async save(query) {
      try {
          error.set('');
          loading.set(true);
          const response = await axios.get(`${BASE_URL}/search.json?title=${query}`);
          loading.set(false);
          set(response.data.docs);
          return response.data;
      } catch(e) {
          loading.set(false);
          set([]);
          error.set(`Error has been occurred. Details: ${e.message}`);
      }
    },

    async delete(query) {
      try {
          error.set('');
          loading.set(true);
          const response = await axios.get(`${BASE_URL}/search.json?title=${query}`);
          loading.set(false);
          set(response.data.docs);
          return response.data;
      } catch(e) {
          loading.set(false);
          set([]);
          error.set(`Error has been occurred. Details: ${e.message}`);
      }
    }
});

export default BooksStore();
