import{d as h,u as P,r as s,c as S,v as T,e as b,b as x,o as I}from"./index-834ac640.js";import{D as k}from"./DataPlaneList-1aa4283e.js";import{u as G}from"./index-0e952743.js";import{Q as f}from"./QueryParameter-70743f73.js";import"./kongponents.es-131ffe48.js";import"./ContentWrapper-be0c9d71.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./DataOverview-9d4155bf.js";import"./DefinitionListItem-e9b95b5e.js";import"./ErrorBlock-42ddf946.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-4325e9aa.js";import"./datadogLogEvents-302eea7b.js";import"./TagList-7389a145.js";import"./store-bb95959d.js";import"./StatusBadge-1d0340ff.js";import"./YamlView.vue_vue_type_script_setup_true_lang-05c499f0.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-82e53469.js";import"./toYaml-4e00099e.js";const X=h({__name:"DataPlaneListView",props:{selectedDppName:{type:String,required:!1,default:null},gatewayType:{type:String,required:!1,default:"true"},offset:{type:Number,required:!1,default:0},isGatewayView:{type:Boolean,required:!1,default:!1}},setup(d){const t=d,v=G(),w=50,g={name:{description:"filter by name or parts of a name"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},_={protocol:{description:"filter by “kuma.io/protocol” value"}},D={},i=P(),l=s([]),p=s(!0),m=s(null),o=s(null),y=s(t.offset),E=S(()=>({...g,...t.isGatewayView?D:_}));T(()=>i.params.mesh,function(){i.name!=="data-plane-list-view"&&i.name!=="gateway-list-view"||u(0)});function F(){const a=f.get("filterFields"),r=a!==null?JSON.parse(a):{};u(t.offset,{...r,gateway:t.gatewayType})}F();async function u(a,r={}){y.value=a,f.set("offset",a>0?a:null),f.set("gatewayType",r.gateway==="true"?"all":r.gateway),p.value=!0;const c=i.params.mesh,n=L(r,w,a,t.isGatewayView);try{const{items:e,next:A}=await v.getAllDataplaneOverviewsFromMesh({mesh:c},n);Array.isArray(e)&&e.length>0?(l.value=e,o.value=A):(l.value=[],o.value=null)}catch(e){e instanceof Error?m.value=e:console.error(e),l.value=[],o.value=null}finally{p.value=!1}}function L(a,r,c,n){const e={...a,size:r,offset:c};return n&&(!("gateway"in e)||e.gateway==="false")?e.gateway="true":n||(e.gateway="false"),e}return(a,r)=>(I(),b(k,{"data-plane-overviews":l.value,"is-loading":p.value,error:m.value,"next-url":o.value,"page-offset":y.value,"selected-dpp-name":t.selectedDppName,"is-gateway-view":t.isGatewayView,"gateway-type":t.gatewayType,"dpp-filter-fields":x(E),onLoadData:u},null,8,["data-plane-overviews","is-loading","error","next-url","page-offset","selected-dpp-name","is-gateway-view","gateway-type","dpp-filter-fields"]))}});export{X as default};
