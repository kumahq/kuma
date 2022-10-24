import{d as y,S as p,o as s,k as _,c as l,H as g,C as m,r as i,g as d,h as v,i as u,w as h,a as E}from"./index.ea0d4a24.js";import{_ as N}from"./CodeBlock.vue_vue_type_style_index_0_lang.f474aa6d.js";import{_ as S}from"./EmptyBlock.vue_vue_type_script_setup_true_lang.b7d1d2f8.js";import{E as k}from"./ErrorBlock.6da0472d.js";import{_ as I}from"./LoadingBlock.vue_vue_type_script_setup_true_lang.c9f72cfb.js";const q={class:"status-info"},z={key:3},P=y({__name:"StatusInfo",props:{isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1},error:{type:[Error,p],required:!1,default:null}},setup(t){return(e,r)=>(s(),_("div",q,[t.isLoading?(s(),l(I,{key:0})):t.hasError?(s(),l(k,{key:1,error:t.error},null,8,["error"])):t.isEmpty?(s(),l(S,{key:2})):(s(),_("div",z,[g(e.$slots,"default",{},void 0,!0)]))]))}});const x=m(P,[["__scopeId","data-v-41cafa0f"]]),D=y({__name:"EnvoyData",props:{dataPath:{type:String,required:!0},queryKey:{type:String,required:!1,default:null},mesh:{type:String,required:!1,default:""},dppName:{type:String,required:!1,default:""},zoneIngressName:{type:String,required:!1,default:""},zoneEgressName:{type:String,required:!1,default:""}},setup(t){const e=t,r=i(!0),o=i(void 0),c=i("");d(()=>e.dppName,function(){n()}),d(()=>e.zoneIngressName,function(){n()}),d(()=>e.zoneEgressName,function(){n()}),v(function(){n()});async function n(){o.value=void 0,r.value=!0;try{let a="";e.mesh!==""&&e.dppName!==""?a=await u.getDataplaneData({dataPath:e.dataPath,mesh:e.mesh,dppName:e.dppName}):e.zoneIngressName!==""?a=await u.getZoneIngressData({dataPath:e.dataPath,zoneIngressName:e.zoneIngressName}):e.zoneEgressName!==""&&(a=await u.getZoneEgressData({dataPath:e.dataPath,zoneEgressName:e.zoneEgressName})),c.value=typeof a=="string"?a:JSON.stringify(a,null,2)}catch(a){a instanceof Error?o.value=a:console.error(a)}finally{r.value=!1}}return(a,B)=>(s(),l(x,{class:"envoy-data","has-error":o.value!==void 0,"is-loading":r.value,error:o.value},{default:h(()=>{var f;return[E(N,{id:`code-block-${t.dataPath}`,language:"json",code:c.value,"is-searchable":"","query-key":(f=t.queryKey)!=null?f:`code-block-${t.dataPath}`},null,8,["id","code","query-key"])]}),_:1},8,["has-error","is-loading","error"]))}});const L=m(D,[["__scopeId","data-v-42af4383"]]);export{L as E,x as S};
