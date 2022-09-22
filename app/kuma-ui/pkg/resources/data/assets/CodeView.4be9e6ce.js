import{_ as C,q as g,I as E,G as v,H as K,r as e,o as t,f as l,b as a,w as n,d,e as h,t as b,n as s,D as x,E as w,h as i}from"./index.2f6d90b0.js";import{C as S,_ as V}from"./CodeBlock.45d0fc63.js";import{_ as L,E as P}from"./ErrorBlock.65239dfd.js";const I={name:"CodeView",components:{CodeBlock:S,EmptyBlock:L,ErrorBlock:P,LoadingBlock:V,KButton:g,KCard:E,KClipboardProvider:v,KPop:K},props:{lang:{type:String,required:!0},copyButtonText:{type:String,default:"Copy to Clipboard"},title:{type:String,default:null},content:{type:String,default:null},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1}}},N=c=>(x("data-v-2801f39f"),c=c(),w(),c),T={class:"code-view"},q={key:3,class:"code-view-content"},D=N(()=>i("div",null,[i("p",null,"Entity copied to clipboard!")],-1));function G(c,H,o,j,z,A){const r=e("LoadingBlock"),_=e("ErrorBlock"),p=e("EmptyBlock"),m=e("CodeBlock"),u=e("KButton"),f=e("KPop"),y=e("KClipboardProvider"),B=e("KCard");return t(),l("div",T,[o.isLoading?(t(),a(r,{key:0})):o.hasError?(t(),a(_,{key:1})):o.isEmpty?(t(),a(p,{key:2})):(t(),l("div",q,[!o.isLoading&&!o.isEmpty?(t(),a(B,{key:0,title:o.title,"border-variant":"noBorder"},{body:n(()=>[d(m,{language:o.lang,code:o.content},null,8,["language","code"])]),actions:n(()=>[o.content?(t(),a(y,{key:0},{default:n(({copyToClipboard:k})=>[d(f,{placement:"bottom"},{content:n(()=>[D]),default:n(()=>[d(u,{appearance:"primary",onClick:()=>{k(o.content)}},{default:n(()=>[h(b(o.copyButtonText),1)]),_:2},1032,["onClick"])]),_:2},1024)]),_:1})):s("",!0)]),_:1},8,["title"])):s("",!0)]))])}const O=C(I,[["render",G],["__scopeId","data-v-2801f39f"]]);export{O as C};
