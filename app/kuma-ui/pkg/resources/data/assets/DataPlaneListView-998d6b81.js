import{d as P,cv as h,bP as s,g as A,h as b,c8 as x,o as I,i as S,u as k}from"./index-08ba2993.js";import{Q as d}from"./QueryParameter-70743f73.js";import{D as G}from"./DataPlaneList-bcca8db2.js";import"./ContentWrapper-fe544c43.js";import"./DataOverview-1eb5b106.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-cf69250c.js";import"./ErrorBlock-21576094.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-778739a1.js";import"./StatusBadge-c118c8ba.js";import"./TagList-e8e9bfa1.js";import"./YamlView.vue_vue_type_script_setup_true_lang-f673f333.js";import"./index-a8834e9c.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-e26b650c.js";import"./_commonjsHelpers-87174ba5.js";const Y=P({__name:"DataPlaneListView",props:{selectedDppName:{type:String,required:!1,default:null},offset:{type:Number,required:!1,default:0},isGatewayView:{type:Boolean,required:!1,default:!1}},setup(v){const t=v,w=50,y={name:{description:"filter by name or parts of a name"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},g={protocol:{description:"filter by “kuma.io/protocol” value"}},_={},i=h(),l=s([]),p=s(!0),f=s(null),o=s(null),m=s(t.offset),D=A(()=>({...y,...t.isGatewayView?_:g}));b(()=>i.params.mesh,function(){i.name!=="data-plane-list-view"&&i.name!=="gateway-list-view"||u(0)});function E(){const a=d.get("filterFields"),r=a!==null?JSON.parse(a):{};u(t.offset,r)}E();async function u(a,r={}){m.value=a,d.set("offset",a>0?a:null),p.value=!0;const c=i.params.mesh,n=F(r,w,a,t.isGatewayView);try{const{items:e,next:L}=await x.getAllDataplaneOverviewsFromMesh({mesh:c},n);Array.isArray(e)&&e.length>0?(l.value=e,o.value=L):(l.value=[],o.value=null)}catch(e){e instanceof Error?f.value=e:console.error(e),l.value=[],o.value=null}finally{p.value=!1}}function F(a,r,c,n){const e={...a,size:r,offset:c};return n&&(!("gateway"in e)||e.gateway==="false")?e.gateway="true":n||(e.gateway="false"),e}return(a,r)=>(I(),S(G,{"data-plane-overviews":l.value,"is-loading":p.value,error:f.value,"next-url":o.value,"page-offset":m.value,"selected-dpp-name":t.selectedDppName,"is-gateway-view":t.isGatewayView,"dpp-filter-fields":k(D),onLoadData:u},null,8,["data-plane-overviews","is-loading","error","next-url","page-offset","selected-dpp-name","is-gateway-view","dpp-filter-fields"]))}});export{Y as default};
