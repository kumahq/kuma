import{d as T,u as A,r as s,c as P,v as S,e as b,b as h,B as x,o as I}from"./index-a24b4f04.js";import{D as k}from"./DataPlaneList-927b9c66.js";import{u as B}from"./index-f7ac63b4.js";import{Q as f}from"./QueryParameter-70743f73.js";import"./kongponents.es-5adaddec.js";import"./ContentWrapper-ddab5a94.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./DataOverview-8bfe7e82.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-8ed542ca.js";import"./ErrorBlock-8e1d70a5.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-4f020979.js";import"./datadogLogEvents-302eea7b.js";import"./TagList-67c1d0cb.js";import"./store-07fabdaf.js";import"./StatusBadge-ab90d1ed.js";import"./YamlView.vue_vue_type_script_setup_true_lang-e4ffe560.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-7b080a3a.js";import"./toYaml-4e00099e.js";const X=T({__name:"DataPlaneListView",props:{selectedDppName:{type:String,required:!1,default:null},gatewayType:{type:String,required:!1,default:"true"},offset:{type:Number,required:!1,default:0},isGatewayView:{type:Boolean,required:!1,default:!1}},setup(y){const t=y,v=B(),w={name:{description:"filter by name or parts of a name"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},g={protocol:{description:"filter by “kuma.io/protocol” value"}},_={},i=A(),l=s([]),n=s(!0),m=s(null),p=s(null),d=s(t.offset),D=P(()=>({...w,...t.isGatewayView?_:g}));S(()=>i.params.mesh,function(){i.name!=="data-plane-list-view"&&i.name!=="gateway-list-view"||u(0)});function E(){const a=f.get("filterFields"),r=a!==null?JSON.parse(a):{};u(t.offset,{...r,gateway:t.gatewayType})}E();async function u(a,r={}){d.value=a,f.set("offset",a>0?a:null),f.set("gatewayType",r.gateway==="true"?"all":r.gateway),n.value=!0;const c=i.params.mesh,o=F(r,x,a,t.isGatewayView);try{const{items:e,next:L}=await v.getAllDataplaneOverviewsFromMesh({mesh:c},o);p.value=L,l.value=e??[]}catch(e){e instanceof Error?m.value=e:console.error(e),l.value=[],p.value=null}finally{n.value=!1}}function F(a,r,c,o){const e={...a,size:r,offset:c};return o&&(!("gateway"in e)||e.gateway==="false")?e.gateway="true":o||(e.gateway="false"),e}return(a,r)=>(I(),b(k,{"data-plane-overviews":l.value,"is-loading":n.value,error:m.value,"next-url":p.value,"page-offset":d.value,"selected-dpp-name":t.selectedDppName,"is-gateway-view":t.isGatewayView,"gateway-type":t.gatewayType,"dpp-filter-fields":h(D),onLoadData:u},null,8,["data-plane-overviews","is-loading","error","next-url","page-offset","selected-dpp-name","is-gateway-view","gateway-type","dpp-filter-fields"]))}});export{X as default};
