import{d as b,u as w,r as s,v as k,o as D,j as E,i as x,g as A,e as T,L as f}from"./index-61cef882.js";import{D as L}from"./DataOverview-ef20b89d.js";import{b as N,u as S}from"./index-01e79acb.js";import{Q as V}from"./QueryParameter-70743f73.js";import"./kongponents.es-d381709c.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-5a7795a6.js";import"./ErrorBlock-e115e1aa.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-fa2a2bb6.js";import"./TagList-25ddbb01.js";import"./StatusBadge-69eda1cc.js";const M={class:"kcard-stack"},z={class:"kcard-border"},G=b({__name:"MeshListView",props:{selectedMeshName:{type:[String,null],required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(v){const i=v,u=N(),h=S(),g={title:u.t("meshes.list.emptyState.title"),message:u.t("meshes.list.emptyState.message")},m=w(),o=s(!0),l=s(null),r=s({headers:[{label:"Name",key:"entity"}],data:[]}),c=s(null),p=s(i.offset);k(()=>m.params.mesh,function(){m.name==="mesh-list-view"&&n(0)}),_();function _(){n(i.offset)}async function n(e){p.value=e,V.set("offset",e>0?e:null),o.value=!0,l.value=null;const a=f;try{const{items:t,next:d}=await h.getAllMeshes({size:a,offset:e});c.value=d,r.value.data=y(t??[])}catch(t){r.value.data=[],t instanceof Error?l.value=t:console.error(t)}finally{o.value=!1}}function y(e){return e.map(a=>{const{name:t}=a;return{entity:a,detailViewRoute:{name:"mesh-detail-view",params:{mesh:t}}}})}return(e,a)=>(D(),E("div",M,[x("div",z,[A(L,{"page-size":T(f),"is-loading":o.value,error:l.value,"empty-state":g,"table-data":r.value,"table-data-is-empty":r.value.data.length===0,next:c.value,"page-offset":p.value,onLoadData:n},null,8,["page-size","is-loading","error","table-data","table-data-is-empty","next","page-offset"])])]))}});export{G as default};
