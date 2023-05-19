import{d as T,u as A,r as s,c as P,v as S,b,G as h,o as x}from"./index-c271a676.js";import{D as I}from"./DataPlaneList-b50c00cb.js";import{u as k}from"./index-0fcc3d29.js";import{Q as m}from"./QueryParameter-70743f73.js";import"./kongponents.es-6cc20401.js";import"./ContentWrapper-794c5fcc.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./DataOverview-3679168e.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-5dc5aad7.js";import"./ErrorBlock-1f74e70f.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-daa04137.js";import"./datadogLogEvents-302eea7b.js";import"./TagList-a690852a.js";import"./store-7a329c21.js";import"./StatusBadge-3e47fb84.js";import"./DefinitionListItem-1834b712.js";import"./YamlView.vue_vue_type_script_setup_true_lang-51a6c3f7.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-7b980fdd.js";import"./toYaml-4e00099e.js";const X=T({__name:"DataPlaneListView",props:{selectedDppName:{type:String,required:!1,default:null},gatewayType:{type:String,required:!1,default:"true"},offset:{type:Number,required:!1,default:0},isGatewayView:{type:Boolean,required:!1,default:!1}},setup(y){const t=y,v=k(),w={name:{description:"filter by name or parts of a name"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},g={protocol:{description:"filter by “kuma.io/protocol” value"}},_={},i=A(),l=s([]),n=s(!0),f=s(null),p=s(null),d=s(t.offset),D=P(()=>({...w,...t.isGatewayView?_:g}));S(()=>i.params.mesh,function(){i.name!=="data-plane-list-view"&&i.name!=="gateway-list-view"||u(0)});function E(){const a=m.get("filterFields"),r=a!==null?JSON.parse(a):{};u(t.offset,{...r,gateway:t.gatewayType})}E();async function u(a,r={}){d.value=a,m.set("offset",a>0?a:null),m.set("gatewayType",r.gateway==="true"?"all":r.gateway),n.value=!0;const c=i.params.mesh,o=F(r,h,a,t.isGatewayView);try{const{items:e,next:L}=await v.getAllDataplaneOverviewsFromMesh({mesh:c},o);p.value=L,l.value=e??[]}catch(e){e instanceof Error?f.value=e:console.error(e),l.value=[],p.value=null}finally{n.value=!1}}function F(a,r,c,o){const e={...a,size:r,offset:c};return o&&(!("gateway"in e)||e.gateway==="false")?e.gateway="true":o||(e.gateway="false"),e}return(a,r)=>(x(),b(I,{"data-plane-overviews":l.value,"is-loading":n.value,error:f.value,"next-url":p.value,"page-offset":d.value,"selected-dpp-name":t.selectedDppName,"is-gateway-view":t.isGatewayView,"gateway-type":t.gatewayType,"dpp-filter-fields":D.value,onLoadData:u},null,8,["data-plane-overviews","is-loading","error","next-url","page-offset","selected-dpp-name","is-gateway-view","gateway-type","dpp-filter-fields"]))}});export{X as default};
