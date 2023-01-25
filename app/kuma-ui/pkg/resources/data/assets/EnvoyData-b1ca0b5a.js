import{A as p,k as i}from"./store-2fff246d.js";import{_ as y}from"./CodeBlock.vue_vue_type_style_index_0_lang-525d6c39.js";import{_ as g}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-4047971f.js";import{E as h}from"./ErrorBlock-a792d9a1.js";import{_}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-d3176fee.js";import{d as m,o as r,h as f,a as l,P as v,r as d,y as u,n as E,w as N,e as k}from"./runtime-dom.esm-bundler-91b41870.js";import{_ as q}from"./_plugin-vue_export-helper-c27b6911.js";const z={key:3},P=m({__name:"StatusInfo",props:{isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1},error:{type:[Error,p],required:!1,default:null}},setup(t){return(e,o)=>(r(),f("div",null,[t.isLoading?(r(),l(_,{key:0})):t.hasError?(r(),l(h,{key:1,error:t.error},null,8,["error"])):t.isEmpty?(r(),l(g,{key:2})):(r(),f("div",z,[v(e.$slots,"default")]))]))}}),S=m({__name:"EnvoyData",props:{dataPath:{type:String,required:!0},queryKey:{type:String,required:!1,default:null},mesh:{type:String,required:!1,default:""},dppName:{type:String,required:!1,default:""},zoneIngressName:{type:String,required:!1,default:""},zoneEgressName:{type:String,required:!1,default:""}},setup(t){const e=t,o=d(!0),s=d(void 0),c=d("");u(()=>e.dppName,function(){n()}),u(()=>e.zoneIngressName,function(){n()}),u(()=>e.zoneEgressName,function(){n()}),E(function(){n()});async function n(){s.value=void 0,o.value=!0;try{let a="";e.mesh!==""&&e.dppName!==""?a=await i.getDataplaneData({dataPath:e.dataPath,mesh:e.mesh,dppName:e.dppName}):e.zoneIngressName!==""?a=await i.getZoneIngressData({dataPath:e.dataPath,zoneIngressName:e.zoneIngressName}):e.zoneEgressName!==""&&(a=await i.getZoneEgressData({dataPath:e.dataPath,zoneEgressName:e.zoneEgressName})),c.value=typeof a=="string"?a:JSON.stringify(a,null,2)}catch(a){a instanceof Error?s.value=a:console.error(a)}finally{o.value=!1}}return(a,I)=>(r(),l(P,{class:"envoy-data","has-error":s.value!==void 0,"is-loading":o.value,error:s.value},{default:N(()=>[k(y,{id:`code-block-${t.dataPath}`,language:"json",code:c.value,"is-searchable":"","query-key":t.queryKey??`code-block-${t.dataPath}`},null,8,["id","code","query-key"])]),_:1},8,["has-error","is-loading","error"]))}});const C=q(S,[["__scopeId","data-v-b9869c64"]]);export{C as E,P as _};
