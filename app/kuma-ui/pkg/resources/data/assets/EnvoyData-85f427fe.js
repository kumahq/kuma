import{d as p,g as f,o as e,l as i,i as s,w as n,p as _,H as m,k as r,a9 as k,E as g,x as v,F as q,j as o,W as x,aq as E,K as S,n as u,s as h,t as b}from"./index-f09cca58.js";import{_ as B}from"./CodeBlock.vue_vue_type_style_index_0_lang-d77f2e48.js";import{_ as C}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-bb9bf655.js";const I={class:"envoy-data-actions"},N=p({__name:"EnvoyData",props:{status:{type:String,required:!0},resource:{type:String,required:!0},src:{type:String,required:!0},queryKey:{type:String,required:!0}},setup(l){const{t:y}=f(),t=l;return(D,K)=>(e(),i("div",null,[t.status!=="online"?(e(),s(r(k),{key:0,appearance:"info"},{alertMessage:n(()=>[_("p",null,m(r(y)("common.detail.no_envoy_data",{resource:t.resource})),1)]),_:1})):(e(),s(h,{key:1,src:t.src},{default:n(({data:a,error:c,refresh:d})=>[c?(e(),s(g,{key:0,error:c},null,8,["error"])):a===void 0?(e(),s(v,{key:1})):a===""?(e(),s(C,{key:2})):(e(),i(q,{key:3},[_("div",I,[o(r(x),{appearance:"primary","data-testid":"envoy-data-refresh-button",onClick:d},{default:n(()=>[o(r(E),{size:r(S)},null,8,["size"]),u(`

            Refresh
          `)]),_:2},1032,["onClick"])]),u(),o(B,{id:"code-block-envoy-data",language:"json",code:typeof a=="string"?a:JSON.stringify(a,null,2),"is-searchable":"","query-key":t.queryKey},null,8,["code","query-key"])],64))]),_:1},8,["src"]))]))}});const w=b(N,[["__scopeId","data-v-faac85b9"]]);export{w as E};
