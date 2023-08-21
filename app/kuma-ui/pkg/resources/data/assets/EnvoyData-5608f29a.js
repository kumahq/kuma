import{d as g,o as r,e as f,a as d,z as _,j as i,I as u,k as h,q as v,h as c,w as p,g as y,b as E,D as N}from"./index-71775e0b.js";import{_ as k}from"./CodeBlock.vue_vue_type_style_index_0_lang-c162bae4.js";import{p as q,E as z,q as I,m as P,f as S}from"./RouteView.vue_vue_type_script_setup_true_lang-8b9f8d0f.js";const B={key:3},D=g({__name:"StatusInfo",props:{isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1},error:{type:[Error,null],required:!1,default:null}},setup(t){return(e,n)=>(r(),f("div",null,[t.isLoading?(r(),d(q,{key:0})):t.hasError||t.error!==null?(r(),d(z,{key:1,error:t.error},null,8,["error"])):t.isEmpty?(r(),d(I,{key:2})):(r(),f("div",B,[_(e.$slots,"default")]))]))}}),$={class:"envoy-data-actions"},b=g({__name:"EnvoyData",props:{dataPath:{type:String,required:!0},queryKey:{type:String,required:!1,default:null},mesh:{type:String,required:!1,default:""},dppName:{type:String,required:!1,default:""},zoneIngressName:{type:String,required:!1,default:""},zoneEgressName:{type:String,required:!1,default:""}},setup(t){const e=t,n=P(),o=i(!0),l=i(null),m=i("");u(()=>e.dppName,function(){s()}),u(()=>e.zoneIngressName,function(){s()}),u(()=>e.zoneEgressName,function(){s()}),h(function(){s()});async function s(){l.value=null,o.value=!0;try{let a="";e.mesh!==""&&e.dppName!==""?a=await n.getDataplaneData({dataPath:e.dataPath,mesh:e.mesh,dppName:e.dppName}):e.zoneIngressName!==""?a=await n.getZoneIngressData({dataPath:e.dataPath,zoneIngressName:e.zoneIngressName}):e.zoneEgressName!==""&&(a=await n.getZoneEgressData({dataPath:e.dataPath,zoneEgressName:e.zoneEgressName})),m.value=typeof a=="string"?a:JSON.stringify(a,null,2)}catch(a){a instanceof Error?l.value=a:console.error(a)}finally{o.value=!1}}return(a,w)=>(r(),f("div",null,[v("div",$,[c(E(N),{disabled:o.value,appearance:"primary",icon:"redo","data-testid":"envoy-data-refresh-button",onClick:s},{default:p(()=>[y(`
        Refresh
      `)]),_:1},8,["disabled"])]),y(),c(D,{"is-loading":o.value,error:l.value},{default:p(()=>[c(k,{id:`code-block-${e.dataPath}`,language:"json",code:m.value,"is-searchable":"","query-key":e.queryKey??`code-block-${e.dataPath}`},null,8,["id","code","query-key"])]),_:1},8,["is-loading","error"])]))}});const L=S(b,[["__scopeId","data-v-afa8dc47"]]);export{L as E,D as _};
