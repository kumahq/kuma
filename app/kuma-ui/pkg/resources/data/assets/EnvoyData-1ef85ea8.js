import{K as E}from"./index-9dd3e7d3.js";import{d as h,l as B,a as i,o as e,c as l,b as s,w as t,p as u,t as C,q as n,F as x,e as r,aq as K,f as m,ar as I,_ as N}from"./index-81fc4a03.js";import{_ as b}from"./CodeBlock.vue_vue_type_style_index_0_lang-5ed6861f.js";import{_ as D}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-10dee361.js";import{E as $}from"./ErrorBlock-358840f7.js";import{_ as S}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-c414e426.js";const V={class:"envoy-data-actions"},w=h({__name:"EnvoyData",props:{status:{},resource:{},src:{},query:{default:""}},emits:["query-change"],setup(p,{emit:d}){const{t:y}=B(),a=p,f=d;return(z,c)=>{const v=i("KAlert"),k=i("KButton");return e(),l("div",null,[a.status!=="online"?(e(),s(v,{key:0,appearance:"info"},{alertMessage:t(()=>[u("p",null,C(n(y)("common.detail.no_envoy_data",{resource:a.resource})),1)]),_:1})):(e(),s(I,{key:1,src:a.src},{default:t(({data:o,error:_,refresh:g})=>[_?(e(),s($,{key:0,error:_},null,8,["error"])):o===void 0?(e(),s(S,{key:1})):o===""?(e(),s(D,{key:2})):(e(),l(x,{key:3},[u("div",V,[r(k,{appearance:"primary","data-testid":"envoy-data-refresh-button",onClick:g},{default:t(()=>[r(n(K),{size:n(E)},null,8,["size"]),m(`

            Refresh
          `)]),_:2},1032,["onClick"])]),m(),r(b,{id:"code-block-envoy-data",language:"json",code:typeof o=="string"?o:JSON.stringify(o,null,2),"is-searchable":"",query:a.query,onQueryChange:c[0]||(c[0]=q=>f("query-change",q))},null,8,["code","query"])],64))]),_:1},8,["src"]))])}}});const M=N(w,[["__scopeId","data-v-065eed8d"]]);export{M as E};
