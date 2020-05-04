<script>
  import ErrorBox from "../components/error_box.svelte";
  import ListItem from "../components/list_find_item.svelte";
  import ListsByMeStore from "../store/lists_by_me";
  const loading = ListsByMeStore.loading;
  const error = ListsByMeStore.error;

  ListsByMeStore.get();

  let find = "all";
  let filterByLabel = "Any label";

  let lists = $ListsByMeStore;
  const foundListTypes = [
    ...new Set($ListsByMeStore.map(item => item.info.type))
  ];

  const defaultListTypes = [
    {
      key: "all",
      description: "Any list type"
    },
    {
      key: "v1",
      description: "free text"
    },
    {
      key: "v2",
      description: "From -> To"
    },
    {
      key: "v4",
      description: "A url and some text"
    },
    {
      key: "v3",
      description: "Concept2 rowing machine log"
    }
  ];

  function getSelectListTypes(lists) {
    const foundListTypes = [...new Set(lists.map(item => item.info.type))];
    const listTypes = [];
    // Add the all option
    listTypes.push(defaultListTypes[0]);
    const filtered = new Set(
      defaultListTypes.filter(item => foundListTypes.includes(item.key))
    );
    filtered.forEach(e => {
      listTypes.push(e);
    });
    return listTypes;
  }

  function getSelectListLabels(lists) {
    const labels = new Set();
    lists.forEach(item => {
      if (item.info.labels.length > 0) {
        item.info.labels.forEach(label => {
          labels.add(label);
        });
      }
    });

    if (labels.size === 0) {
      return [];
    }

    return ["Any label", ...labels];
  }

  function filterListsByFilters(lists, find, filterByLabel) {
    let filtered = lists.filter(item => {
      if (find == "all") {
        return true;
      }
      return find == item.info.type;
    });

    filtered = filtered.filter(item => {
      if (filterByLabel == "Any label") {
        return true;
      }
      return item.info.labels.includes(filterByLabel);
    });

    return filtered;
  }

  function hasLabels(listLabels) {
    return !!listLabels.length;
  }

  function reset() {
    find = "all";
    filterByLabel = "Any label";
  }

  $: filterLists = filterListsByFilters($ListsByMeStore, find, filterByLabel);
  $: listTypes = getSelectListTypes(filterLists);
  $: listLabels = getSelectListLabels(filterLists);
</script>

<div class="pa3 pa2-ns">

  <div class="pl0 measure center">
    {#if $error}
      error is {$error}
    {:else if $loading}
      Loading...
    {:else}
      <div>
        <fieldset class="bn">
          <div class="flex items-center mb2">
            <span>Filter</span>
          </div>

          <div class="flex items-center mb2">
            <select bind:value={find}>
              {#each listTypes as listType}
                <option value={listType.key}>{listType.description}</option>
              {/each}
            </select>
          </div>

          {#if hasLabels(listLabels)}
            <div class="flex items-center mb2">
              <select bind:value={filterByLabel}>
                {#each listLabels as label}
                  <option value={label}>{label}</option>
                {/each}
              </select>
            </div>
          {/if}

          <div class="flex items-center mb2">
            <button on:click={reset}>reset</button>
          </div>
        </fieldset>
      </div>
      <ul class="list pl0 measure center">
        {#each filterLists as aList}
          <ListItem title={aList.info.title} uuid={aList.uuid} />
        {/each}
      </ul>
    {/if}
  </div>
</div>
