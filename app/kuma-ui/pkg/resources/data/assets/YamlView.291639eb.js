import{j as y}from"./index.58caa11d.js";import{d as _,r as B,f as E,o as s,j as k,c as n,w as b,a as g,u as o,R as V,q as j,D as w}from"./index.3a3d021f.js";import{_ as v}from"./CodeBlock.vue_vue_type_style_index_0_lang.1ff82bc9.js";import{_ as S}from"./EmptyBlock.vue_vue_type_script_setup_true_lang.ed08f409.js";import{E as H}from"./ErrorBlock.e950a812.js";import{_ as Y}from"./LoadingBlock.vue_vue_type_script_setup_true_lang.fa633f42.js";const C={class:"yaml-view"},M={key:3,class:"yaml-view-content"},O=_({__name:"YamlView",props:{id:{type:String,required:!0},content:{type:Object,required:!1,default:null},isLoading:{type:Boolean,required:!1,default:!1},hasError:{type:Boolean,required:!1,default:!1},isEmpty:{type:Boolean,required:!1,default:!1},codeMaxHeight:{type:String,required:!1,default:null},isSearchable:{type:Boolean,required:!1,default:!1}},setup(e){const a=e,p=j(),c=[{hash:"#universal",title:"Universal"},{hash:"#kubernetes",title:"Kubernetes"}],i=B(c[0].hash),l=p.getters["config/getEnvironment"];typeof l=="string"&&(i.value="#"+l);const m=E(()=>{var f;const t={};if(t.apiVersion="kuma.io/v1alpha1",t.kind=a.content.type,a.content.mesh!==void 0&&(t.mesh=a.content.mesh),(f=a.content.name)!=null&&f.includes(".")){const h=a.content.name.split("."),q=h.pop(),x=h.join(".");t.metadata={name:x,namespace:q}}else t.metadata={name:a.content.name};const{type:r,name:d,mesh:$,...u}=a.content;return Object.keys(u).length>0&&(t.spec=u),{universal:y(a.content),kubernetes:y(t)}});return(t,r)=>(s(),k("div",C,[e.isLoading?(s(),n(Y,{key:0})):e.hasError?(s(),n(H,{key:1})):e.isEmpty?(s(),n(S,{key:2})):(s(),k("div",M,[(s(),n(o(V),{key:o(l),modelValue:i.value,"onUpdate:modelValue":r[0]||(r[0]=d=>i.value=d),tabs:c},{universal:b(()=>[g(v,{id:e.id,language:"yaml",code:o(m).universal,"is-searchable":e.isSearchable,"query-key":e.id,"code-max-height":e.codeMaxHeight},null,8,["id","code","is-searchable","query-key","code-max-height"])]),kubernetes:b(()=>[g(v,{id:e.id,language:"yaml",code:o(m).kubernetes,"is-searchable":e.isSearchable,"query-key":e.id,"code-max-height":e.codeMaxHeight},null,8,["id","code","is-searchable","query-key","code-max-height"])]),_:1},8,["modelValue"]))]))]))}});const R=w(O,[["__scopeId","data-v-f92420cb"]]);export{R as Y};
