import{d as S,I as x,F as N,o as u,m as y,w as i,c as w,G as C,E as I,e as b,a as m,b as r,k as s,t as c,H as L,J as V,l as k,K as $,p as q,q as O}from"./index-B_icS-nL.js";import{O as B,a as E,b as R}from"./OnboardingPage-BjFVU4-F.js";const A=["aria-hidden"],z='<path d="M9.7 18.025L4 12.325L5.425 10.9L9.7 15.175L18.875 6L20.3 7.425L9.7 18.025Z" fill="currentColor"/>',M=S({__name:"CheckIcon",props:{title:{type:String,required:!1,default:""},color:{type:String,required:!1,default:"currentColor"},display:{type:String,required:!1,default:"block"},decorative:{type:Boolean,required:!1,default:!1},size:{type:[Number,String],required:!1,default:x,validator:a=>{if(typeof a=="number"&&a>0)return!0;if(typeof a=="string"){const t=String(a).replace(/px/gi,""),e=Number(t);if(e&&!isNaN(e)&&Number.isInteger(e)&&e>0)return!0}return!1}},as:{type:String,required:!1,default:"span"},staticIds:{type:Boolean,default:!1}},setup(a){const t=a,e=N(()=>{if(typeof t.size=="number"&&t.size>0)return`${t.size}px`;if(typeof t.size=="string"){const o=String(t.size).replace(/px/gi,""),n=Number(o);if(n&&!isNaN(n)&&Number.isInteger(n)&&n>0)return`${n}px`}return x}),p=N(()=>({boxSizing:"border-box",color:t.color,display:t.display,flexShrink:"0",height:e.value,lineHeight:"0",width:e.value,pointerEvents:t.decorative?"none":void 0})),g=o=>{const n={},l=Math.random().toString(36).substring(2,12);return o.replace(/id="([^"]+)"/g,(_,d)=>{const h=`${l}-${d}`;return n[d]=h,`id="${h}"`}).replace(/#([^\s^")]+)/g,(_,d)=>n[d]?`#${n[d]}`:_)},f=o=>{const n={"<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#039;","`":"&#039;"};return o.replace(/[<>"'`]/g,l=>n[l])},v=`${t.title?`<title data-testid="kui-icon-svg-title">${f(t.title)}</title>`:""}${t.staticIds?z:g(z)}`;return(o,n)=>(u(),y(I(a.as),{"aria-hidden":a.decorative?"true":void 0,class:"kui-icon check-icon","data-testid":"kui-icon-wrapper-check-icon",style:C(p.value),tabindex:a.decorative?"-1":void 0},{default:i(()=>[(u(),w("svg",{"aria-hidden":a.decorative?"true":void 0,"data-testid":"kui-icon-svg-check-icon",fill:"none",height:"100%",role:"img",viewBox:"0 0 24 24",width:"100%",xmlns:"http://www.w3.org/2000/svg",innerHTML:v},null,8,A))]),_:1},8,["aria-hidden","style","tabindex"]))}}),T={"data-testid":"kuma-environment"},H={class:"item-status-list-wrapper"},W={class:"item-status-list"},K={class:"circle mr-2"},D=S({__name:"OnboardingWelcomeView",setup(a){return(t,e)=>{const p=b("RouteTitle"),g=b("AppView"),f=b("RouteView");return u(),y(f,{name:"onboarding-welcome-view"},{default:i(({env:v,t:o,can:n})=>[m(p,{title:o("onboarding.routes.welcome.title",{name:o("common.product.name")}),render:!1},null,8,["title"]),e[10]||(e[10]=r()),m(g,null,{default:i(()=>[s("div",null,[m(B,null,{header:i(()=>[m(E,null,{title:i(()=>[r(`
                Welcome to `+c(o("common.product.name")),1)]),description:i(()=>[s("p",null,[r(`
                  Congratulations on downloading `+c(o("common.product.name"))+"! You are just a ",1),e[0]||(e[0]=s("strong",null,"few minutes",-1)),e[1]||(e[1]=r(` away from getting your service mesh fully online.
                `))]),e[4]||(e[4]=r()),s("p",null,[e[2]||(e[2]=r(`
                  We have automatically detected that you are running on `)),s("strong",T,c(o(`common.product.environment.${v("KUMA_ENVIRONMENT")}`)),1),e[3]||(e[3]=r(`.
                `))])]),_:2},1024)]),content:i(()=>[e[6]||(e[6]=s("h2",{class:"text-center"},`
              Let’s get started:
            `,-1)),e[7]||(e[7]=r()),s("div",H,[s("ul",W,[(u(!0),w(L,null,V([{name:`Run ${o("common.product.name")} control plane`,status:!0},{name:"Learn about deployments",status:!1},{name:"Learn about configuration storage",status:!1},...n("use zones")?[{name:"Add zones",status:!1}]:[],{name:"Create the mesh",status:!1},{name:"Add services",status:!1},{name:"Go to the dashboard",status:!1}],l=>(u(),w("li",{key:l.name},[s("span",K,[l.status?(u(),y(k(M),{key:0,size:k($)},null,8,["size"])):q("",!0)]),r(" "+c(l.name),1)]))),128))])])]),navigation:i(()=>[m(R,{"next-step":"onboarding-deployment-types-view"})]),_:2},1024)])]),_:2},1024)]),_:1})}}}),U=O(D,[["__scopeId","data-v-e0b0cf7b"]]);export{U as default};
